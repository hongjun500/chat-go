package protocol

import (
	"fmt"
	"io"
)

// Encoding 表示消息负载的编码方式
type Encoding string

const (
	EncodingJSON     Encoding = Json
	EncodingProtobuf Encoding = Protobuf
)

// MessageType 表示系统支持的业务消息类型
type MessageType string

const (
	MsgText      MessageType = "text"
	MsgCommand   MessageType = "command"
	MsgFileMeta  MessageType = "file_meta"
	MsgFileChunk MessageType = "file_chunk"
	MsgAck       MessageType = "ack"
	MsgPing      MessageType = "ping"
	MsgPong      MessageType = "pong"
)

// Manager 协议管理器，负责协议层的核心功能
type Manager struct {
	codec   MessageCodec
	factory *MessageFactory
	router  *MessageRouter
}

// NewProtocolManager 创建协议管理器
func NewProtocolManager(codecType int) *Manager {
	codec, err := NewCodec(codecType)
	if err != nil {
		// 如果创建失败，回退到 JSON 编解码器
		codec = &JSONCodec{}
	}
	return &Manager{
		codec:   codec,
		factory: NewMessageFactory(),
		router:  NewMessageRouter(),
	}
}

// GetCodec 获取编解码器
func (p *Manager) GetCodec() MessageCodec {
	return p.codec
}

// GetMessageFactory 获取消息工厂
func (p *Manager) GetMessageFactory() *MessageFactory {
	return p.factory
}

// GetRouter 获取消息路由器
func (p *Manager) GetRouter() *MessageRouter {
	return p.router
}

// EncodeMessage 编码消息
func (p *Manager) EncodeMessage(w io.Writer, envelope *Envelope) error {
	return p.codec.Encode(w, envelope)
}

// DecodeMessage 解码消息
func (p *Manager) DecodeMessage(r io.Reader, envelope *Envelope, maxSize int) error {
	return p.codec.Decode(r, envelope, maxSize)
}

// ProcessMessage 处理消息（解码+路由）
func (p *Manager) ProcessMessage(r io.Reader, maxSize int) error {
	var envelope Envelope
	if err := p.DecodeMessage(r, &envelope, maxSize); err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	return p.router.Dispatch(&envelope)
}

// RegisterMessageHandler 注册消息处理器
func (p *Manager) RegisterMessageHandler(msgType MessageType, handler MessageHandler) {
	p.router.RegisterHandler(msgType, handler)
}

// SetDefaultMessageHandler 设置默认消息处理器
func (p *Manager) SetDefaultMessageHandler(handler MessageHandler) {
	p.router.SetDefaultHandler(handler)
}
