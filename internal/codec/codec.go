package codec

import (
	"fmt"
	"github.com/hongjun500/chat-go/internal/protocol"
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
	Encode(w io.Writer, m *protocol.Envelope) error
	Decode(r io.Reader, m *protocol.Envelope, maxSize int) error
}

// NewCodec 根据编码类型创建相应的编解码器
func NewCodec(codecType string) (MessageCodec, error) {
	switch codecType {
	case "json":
		return &JSONCodec{}, nil
	case "protobuf":
		return &ProtobufCodec{}, nil
	default:
		return nil, fmt.Errorf("unsupported codec type: %s", codecType)
	}
}
