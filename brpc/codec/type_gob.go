package codec

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	// buf 考虑提供快捷返回的方法，在 enc 过程中只要成功把数据写入 buf 中就予以响应成功
	buf *bufio.Writer
	dec *gob.Decoder
	enc *gob.Encoder
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

func (c *GobCodec) ReadHeader(h *Header) error {
	err := c.dec.Decode(h)
	log.Printf("gob codec info: reading header, codec:%#v\n", c.conn)
	return err
}

func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *GobCodec) Write(h *Header, body any) error {
	var err error
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(h)
	if err = c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}

	_ = gob.NewEncoder(&buf).Encode(body)

	// fmt.Printf("原始内容：%#v, %#v, gob encode result: %v\n\n", h, body, buf.Bytes())

	if err = c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}

func (c *GobCodec) GetCC() io.ReadWriteCloser {
	return c.conn
}
