package werpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
	"werpc/codec"
)

type Call struct {
	Seq           uint64
	ServiceMethod string      //format: "Service.Method"
	Args          interface{} // arguments to the function
	Reply         interface{} // reply from the function
	Error         error       // if error occurs, it will be set
	Done          chan *Call  // Strobes when call is complete.
}

func (call *Call) done() {
	//为了支持异步调用，Call 结构体中添加了一个字段 Done
	//Done 的类型是 chan *Call，当调用结束时，会调用 call.done() 通知调用方。
	call.Done <- call
}

type Client struct {
	cc      codec.Codec //消息编解码器
	opt     *Option
	sending sync.Mutex       //互斥锁，和服务端类似，为了保证请求的有序发送，即防止出现多个请求报文混淆。
	header  codec.Header     //每个请求的消息头，header 只有在请求发送时才需要，而请求发送是互斥的，因此每个客户端只需要一个
	mu      sync.Mutex       //防止并发修改 pending closing 和 shutdown
	seq     uint64           //用于给发送的请求编号，每个请求拥有唯一编号。
	pending map[uint64]*Call //存储未处理完的请求，键是编号，值是 Call 实例
	//closing 和 shutdown 任意一个值置为 true，则表示 Client 处于不可用的状态，但有些许的差别
	//closing 是用户主动关闭的，即调用 Close 方法，而 shutdown 置为 true 一般是有错误发生。
	closing  bool
	shutdown bool
}

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	//创建 Client 实例时，首先需要完成一开始的协议交换，即发送 Option 信息给服务端。
	//协商好消息的编解码方式之后，再创建一个子协程调用 receive() 接收响应
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}

	// send options with server
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}

	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, // seq starts with 1, 0 means invalid call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

//parseOptions 方便用户自定义 Option ,接受一个或0个 *Option 或 nil 值
func parseOptions(opts ...*Option) (*Option, error) {
	// if opts is nil or pass nil as parameter, use DefaultOption
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

//Dial 连接到指定网络地址上的RPC服务器
func Dial(network, address string, opts ...*Option) (*Client, error) {
	return dialTimeout(NewClient, network, address, opts...)
}

// Close the connection
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

// IsAvailable return true if the client does work
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connect is shutdown")

//registerCall 将参数 call 添加到 client.pending 中，并更新 client.seq
func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing || client.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

//removeCall 根据 seq，从 client.pending 中移除对应的 call，并返回call
func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

//terminateCalls 服务端或客户端发生错误时调用，将 shutdown 设置为 true，且将错误信息通知所有 pending 状态的 call
func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

func (client *Client) send(call *Call) {
	// make sure that the client will send a complete request
	client.sending.Lock()
	defer client.sending.Unlock()

	// register this call.
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// prepare request header
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	// encode and send the request
	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		// call may be nil, it usually means that Write partially failed,
		// client has received the response and handled
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

// Go invokes the function asynchronously.
// It returns the Call structure representing the invocation.
func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	// Go 是一个异步接口，返回 call 实例
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	client.send(call)
	return call
}

// Call invokes the named function, waits for it to complete,
// and returns its error status.
// 	可使用 context.WithTimeout 创建具备超时检测能力的 context 对象来控制
//	e.g:
//			ctx, _ := context.WithTimeout(context.Background(), time.Second)
//			var reply int
//			err := client.Call(ctx, "Foo.Sum", &Args{1, 2}, &reply)
func (client *Client) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	//Call 是对 Go 的封装，阻塞 call.Done，等待响应返回，是一个同步接口
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
}

//receive 接受服务返回的响应
func (client *Client) receive() {
	var err error
	for err == nil {
		//读取header
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}
		//已返回响应,获取响应对应的请求并移除请求
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			//call 不存在，可能是请求没有发送完整，或者因为其他原因被取消，但是服务端仍旧处理了
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			//表示服务端发生错误
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			//通知 client 该请求结束
			call.done()
		default:
			//call存在,服务器正常,则从body中读取Reply的值
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	// error occurs, so terminateCalls pending calls
	client.terminateCalls(err)
}

type clientResult struct {
	client *Client
	err    error
}

type newClientFunc func(conn net.Conn, opt *Option) (client *Client, err error)

func dialTimeout(f newClientFunc, network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	// close the connection if client is nil
	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()
	ch := make(chan clientResult)
	go func() {
		client, err := f(conn, opt)
		ch <- clientResult{client: client, err: err}
	}()
	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case result := <-ch:
		return result.client, result.err
	}
}
