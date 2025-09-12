package protocol

import (
	"encoding/json"
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

// Welcome 创建欢迎消息
func (p *Protocol) Welcome(text string) *Envelope {
	text = strings.TrimSpace(text)
	return &Envelope{
		Type:    MsgText,
		Ts:      time.Now().UnixMilli(),
		Data:    []byte(text),
		Version: "1.0",
	}
}

// CreateTextMessage 创建文本消息
func (p *Protocol) CreateTextMessage(text string) *Envelope {
	return &Envelope{
		Type:    MsgText,
		Ts:      time.Now().UnixMilli(),
		Data:    []byte(text),
		Version: "1.0",
	}
}

// CreateAckMessage 创建确认消息
func (p *Protocol) CreateAckMessage(status string) *Envelope {
	payload := AckPayload{Status: status}
	data, _ := p.codec.(*JSONCodec) // 简化处理，实际可以更严谨
	if data != nil {
		// 如果是 JSON 编解码器，直接序列化
		if jsonData, err := json.Marshal(payload); err == nil {
			return &Envelope{
				Type:    MsgAck,
				Ts:      time.Now().UnixMilli(),
				Data:    jsonData,
				Version: "1.0",
			}
		}
	}
	// 回退到简单的字符串格式
	return &Envelope{
		Type:    MsgAck,
		Ts:      time.Now().UnixMilli(),
		Data:    []byte(status),
		Version: "1.0",
	}
}
