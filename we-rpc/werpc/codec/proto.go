package codec

import (
	"bufio"
	"io"
)

type ProtoCodec struct {
	conn io.ReadWriteCloser //由构建函数传入，通常是通过 TCP 或者 Unix 建立 socket 时得到的链接实例
	buf  *bufio.Writer      //为了防止阻塞而创建的带缓冲的Writer
}

var _ Codec = (*ProtoCodec)(nil)

//TODO 完成proto方式编码

func NewProtoCodec(conn io.ReadWriteCloser) *ProtoCodec {
	return &ProtoCodec{conn: conn}
}

func (c *ProtoCodec) Close() error {
	return c.conn.Close()
}

func (c *ProtoCodec) ReadHeader(header *Header) error {
	panic("implement me")
}

func (c *ProtoCodec) ReadBody(body interface{}) error {
	panic("implement me")
}

func (c *ProtoCodec) Write(header *Header, body interface{}) error {
	panic("implement me")
}
