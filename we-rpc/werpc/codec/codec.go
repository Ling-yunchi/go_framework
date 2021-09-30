package codec

import (
	"io"
)

type Header struct {
	ServiceMethod string //format: "Service.Method"
	Seq           uint64 //请求序号,可以认为是某个请求的id,用于区分不同请求
	Error         string //错误信息，客户端置为空，服务端如果如果发生错误，将错误信息置于 Error 中
}

type Codec interface {
	io.Closer
	ReadHeader(header *Header) error
	ReadBody(body interface{}) error
	Write(header *Header, body interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType   Type = "application/gob"
	ProtoType Type = "application/proto"
	JsonType  Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
	NewCodecFuncMap[JsonType] = NewJsonCodec
}
