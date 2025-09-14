package protocol

import (
	"strings"
	"time"
)

// Protocol 负责消息创建和编解码管理
type Protocol struct {
	codec MessageCodec
}

// NewProtocol 创建新的协议实例
func NewProtocol(codecType int) *Protocol {
	codec, err := NewCodec(codecType)
	if err != nil {
		// 如果创建失败，回退到 JSON 编解码器
		codec = &JSONCodec{}
	}
	return &Protocol{
		codec: codec,
	}
}

// GetCodec 获取编解码器
func (p *Protocol) GetCodec() MessageCodec {
	return p.codec
}

// SetCodec 设置编解码器
func (p *Protocol) SetCodec(codec MessageCodec) {
	p.codec = codec
}

// CreateTextMessage 创建文本消息
func (p *Protocol) CreateTextMessage(text string) *Envelope {
	return &Envelope{
		Type:    MsgText,
		Ts:      time.Now().UnixMilli(),
		Data:    []byte(strings.TrimSpace(text)),
		Version: "1.0",
	}
}

// CreateAckMessage 创建确认消息
func (p *Protocol) CreateAckMessage(status string) *Envelope {
	return &Envelope{
		Version: "1.0",
		Type:    MsgAck,
		Ts:      time.Now().UnixMilli(),
		Data:    []byte(status),
	}
}
