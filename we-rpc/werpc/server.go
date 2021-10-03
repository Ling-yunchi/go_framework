package werpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
	"werpc/codec"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber    int        //MagicNumber 标识这是一个WeRPC请求
	CodecType      codec.Type //客户端可能选择不同的方式来编码body
	ConnectTimeout time.Duration
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10,
}

//涉及协议协商的这部分信息，需要设计固定的字节来传输的。
//但是为了实现上更简单，WeRPC 客户端固定采用 GoBinary 编码 Option
//后续的 header 和 body 的编码方式由 Option 中的 CodeType 指定
//服务端首先使用 JSON 解码 Option，然后通过 Option 的 CodeType 解码剩余的内容
//	| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
//	| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定     ------->|
//在一次连接中，Option 固定在报文的最开始，Header 和 Body 可以有多个
//	| Option | Header1 | Body1 | Header2 | Body2 | ...

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

//DefaultServer 一个默认的Server实例,方便用户使用
var DefaultServer = NewServer()

//Accept 接受listener上每一个传入的连接并为请求提供服务
func (server *Server) Accept(listener net.Listener) {
	//死循环等待socket连接建立
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		//开启子协程处理
		go server.ServeConn(conn)
	}
}

//Accept 接受listener上每一个传入的连接并为请求提供服务
func Accept(listener net.Listener) {
	DefaultServer.Accept(listener)
}

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	//使用 json.NewDecoder 反序列化得到 Option 实例,检查 MagicNumber 和 CodeType 的值是否正确
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}

	//根据 CodeType 得到对应的消息编解码器
	codecFunc := codec.NewCodecFuncMap[opt.CodecType]
	if codecFunc == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	//转交给serveCodec处理
	server.serveCodec(codecFunc(conn), &opt)
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec, opt *Option) {
	sending := new(sync.Mutex) //确保发送完整的response
	wg := new(sync.WaitGroup)  //等待所有请求被处理
	//在一次连接中，允许接收多个请求,使用for不断循环等待请求
	for {
		//读取请求
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break //不可能恢复,关闭连接
			}
			//回复请求
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		//处理请求
		wg.Add(1)
		//使用协程并发执行请求
		go server.handleRequest(cc, req, sending, wg, opt.HandleTimeout)
	}
	wg.Wait()
	_ = cc.Close()
}

type request struct {
	h                    *codec.Header
	argValue, replyValue reflect.Value //  请求参数类型与返回值类型
	mtype                *methodType
	svc                  *service
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var header codec.Header
	if err := cc.ReadHeader(&header); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &header, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	header, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: header}
	req.svc, req.mtype, err = server.findService(header.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argValue = req.mtype.newArgv()
	req.replyValue = req.mtype.newReplyv()

	// make sure that argvi is a pointer, ReadBody need a pointer as parameter
	argvi := req.argValue.Interface()
	if req.argValue.Type().Kind() != reflect.Ptr {
		argvi = req.argValue.Addr().Interface()
	}
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, header *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(header, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	//处理请求是并发的，但是回复请求的报文必须是逐个发送的，并发容易导致多个回复报文交织在一起，客户端无法解析。在这里使用锁(sending)保证
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.svc.call(req.mtype, req.argValue, req.replyValue)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		server.sendResponse(cc, req.h, req.replyValue.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	case <-time.After(timeout):
		//time.After() 先于 called 接收到消息，说明处理已经超时，called 和 sent 都将被阻塞。
		//在 case <-time.After(timeout) 处调用 sendResponse
		req.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		//called 信道接收到消息，代表处理没有超时，继续执行 sendResponse
		<-sent
	}
}

func (server *Server) Register(rcvr interface{}) error {
	service := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(service.name, service); dup {
		return errors.New("rpc: service already defined: " + service.name)
	}
	return nil
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}

//支持HTTP协议,实现代理服务器对请求的转发

const (
	connected        = "200 Connected to We RPC"
	defaultRPCPath   = "/_werpc_"
	defaultDebugPath = "/debug/werpc"
)

func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must CONNECT\n")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}

	// "HTTP/1.0 "后面少了一个空格导致我debug了一上午qwqqqq
	_, _ = io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.ServeConn(conn)
}

// HandleHTTP registers an HTTP handler for RPC messages on rpcPath.
// It is still necessary to invoke http.Serve(), typically in a go statement.
func (server *Server) HandleHTTP() {
	http.Handle(defaultRPCPath, server)
	//debug page
	http.Handle(defaultDebugPath, debugHTTP{server})
	log.Println("rpc server debug path:", defaultDebugPath)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
