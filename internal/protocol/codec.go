package protocol

import (
	"fmt"
	"io"
)

const (
	CodecJson = iota
	CodecProtobuf
	CodecMsgpack
)

const (
	Json     = "json"
	Protobuf = "protobuf"
	Msgpack  = "msgpack"
)

var codecFactories = map[int]func() MessageCodec{
	CodecJson:     func() MessageCodec { return &JSONCodec{} },
	CodecProtobuf: func() MessageCodec { return &ProtobufCodec{} },
}

// MessageCodec 消息体数据编码解码器
type MessageCodec interface {
	Name() string
	Encode(w io.Writer, m *Envelope) error
	Decode(r io.Reader, m *Envelope, maxSize int) error
}

// NewCodec 根据编码类型创建相应的编解码器
func NewCodec(cc int) (MessageCodec, error) {
	if factory, ok := codecFactories[cc]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("unsupported codec type: %d", cc)
}
