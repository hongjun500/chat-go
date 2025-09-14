package protocol

import (
	"fmt"
	"io"
)

// ProtocolManager 协议管理器，负责协议层的核心功能
type ProtocolManager struct {
	codec   MessageCodec
	factory *MessageFactory
	router  *MessageRouter
}

// NewProtocolManager 创建协议管理器
func NewProtocolManager(codecType int) *ProtocolManager {
	codec, err := NewCodec(codecType)
	if err != nil {
		// 如果创建失败，回退到 JSON 编解码器
		codec = &JSONCodec{}
	}
	
	return &ProtocolManager{
		codec:   codec,
		factory: NewMessageFactory(),
		router:  NewMessageRouter(),
	}
}

// GetCodec 获取编解码器
func (p *ProtocolManager) GetCodec() MessageCodec {
	return p.codec
}

// SetCodec 设置编解码器
func (p *ProtocolManager) SetCodec(codec MessageCodec) {
	p.codec = codec
}

// GetMessageFactory 获取消息工厂
func (p *ProtocolManager) GetMessageFactory() *MessageFactory {
	return p.factory
}

// GetRouter 获取消息路由器
func (p *ProtocolManager) GetRouter() *MessageRouter {
	return p.router
}

// EncodeMessage 编码消息
func (p *ProtocolManager) EncodeMessage(w io.Writer, envelope *Envelope) error {
	return p.codec.Encode(w, envelope)
}

// DecodeMessage 解码消息
func (p *ProtocolManager) DecodeMessage(r io.Reader, envelope *Envelope, maxSize int) error {
	return p.codec.Decode(r, envelope, maxSize)
}

// ProcessMessage 处理消息（解码+路由）
func (p *ProtocolManager) ProcessMessage(r io.Reader, maxSize int) error {
	var envelope Envelope
	if err := p.DecodeMessage(r, &envelope, maxSize); err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}
	
	return p.router.Dispatch(&envelope)
}

// RegisterMessageHandler 注册消息处理器
func (p *ProtocolManager) RegisterMessageHandler(msgType MessageType, handler MessageHandler) {
	p.router.RegisterHandler(msgType, handler)
}

// SetDefaultMessageHandler 设置默认消息处理器
func (p *ProtocolManager) SetDefaultMessageHandler(handler MessageHandler) {
	p.router.SetDefaultHandler(handler)
}

// ---- 向后兼容的 Protocol 结构 ----

// Protocol 保持向后兼容的协议结构（已废弃，建议使用 ProtocolManager）
// Deprecated: 使用 ProtocolManager 替代
type Protocol struct {
	manager *ProtocolManager
}

// NewProtocol 创建协议实例（向后兼容）
// Deprecated: 使用 NewProtocolManager 替代
func NewProtocol(codecType int) *Protocol {
	return &Protocol{
		manager: NewProtocolManager(codecType),
	}
}

// GetCodec 获取编解码器（向后兼容）
func (p *Protocol) GetCodec() MessageCodec {
	return p.manager.GetCodec()
}

// SetCodec 设置编解码器（向后兼容）
func (p *Protocol) SetCodec(codec MessageCodec) {
	p.manager.SetCodec(codec)
}

// CreateTextMessage 创建文本消息（向后兼容，简化版本）
func (p *Protocol) CreateTextMessage(text string) *Envelope {
	return p.manager.GetMessageFactory().CreateTextMessage(text)
}

// CreateAckMessage 创建确认消息（向后兼容）
func (p *Protocol) CreateAckMessage(status string) *Envelope {
	return p.manager.GetMessageFactory().CreateAckMessage(status, "")
}
