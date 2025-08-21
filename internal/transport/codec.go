package transport

import (
	"io"
)

const (
	ApplicationJson     = "application/json"
	ApplicationProtobuf = "application/x-protobuf"
)

type ContentType string

// MessageCodec 消息体数据编码解码器
type MessageCodec interface {
	ContentType() string
	Encode(w io.Writer, m *Envelope) error
	Decode(r io.Reader, m *Envelope, maxSize int) error
}
