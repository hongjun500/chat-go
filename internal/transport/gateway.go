package transport

import (
	"github.com/hongjun500/chat-go/internal/protocol"
)

type handlerFunc func(Session, *protocol.Envelope)

// SimpleGateway 简单的网关实现，专注于消息转发和会话管理
// 作为传输层与业务层的桥梁，不直接处理协议编解码
type SimpleGateway struct {
	sessionManager  *SessionManager
	messageHandlers map[string]handlerFunc
	protocolManager *protocol.Manager
}

// NewSimpleGateway 创建简单网关
func NewSimpleGateway(codecType int) *SimpleGateway {
	return &SimpleGateway{
		sessionManager:  NewSessionManager(),
		messageHandlers: make(map[string]handlerFunc),
		protocolManager: protocol.NewProtocolManager(codecType),
	}
}

// RegisterHandler 注册会话级别的处理器
func (g *SimpleGateway) RegisterHandler(msgType string, handler handlerFunc) {
	g.messageHandlers[msgType] = handler
}

// OnSessionOpen 会话开启事件
func (g *SimpleGateway) OnSessionOpen(sess Session) {
	g.sessionManager.AddSession(sess)
	// 发送欢迎消息或进行会话初始化
	g.sendWelcomeMessage(sess)
}

// OnEnvelope 处理收到的消息
func (g *SimpleGateway) OnEnvelope(sess Session, msg *protocol.Envelope) {
	// 首先尝试会话级别的处理器
	if handler, exists := g.messageHandlers[string(msg.Type)]; exists {
		handler(sess, msg)
		return
	}

	// 回退到协议层的消息路由器
	if err := g.protocolManager.GetRouter().Dispatch(msg); err != nil {
		// 如果无法处理，发送错误确认
		ackMsg := g.protocolManager.GetMessageFactory().CreateAckMessage("unknown_type", msg.Mid)
		_ = sess.SendEnvelope(ackMsg)
	}
}

// OnSessionClose 会话关闭事件
func (g *SimpleGateway) OnSessionClose(sess Session) {
	g.sessionManager.RemoveSession(sess.ID())
}

// GetSessionManager 获取会话管理器
func (g *SimpleGateway) GetSessionManager() *SessionManager {
	return g.sessionManager
}

// GetProtocolManager 获取协议管理器
func (g *SimpleGateway) GetProtocolManager() *protocol.Manager {
	return g.protocolManager
}

// RegisterProtocolHandler 注册协议层消息处理器
func (g *SimpleGateway) RegisterProtocolHandler(msgType protocol.MessageType, handler protocol.MessageHandler) {
	g.protocolManager.RegisterMessageHandler(msgType, handler)
}

// BroadcastMessage 向所有会话广播消息
func (g *SimpleGateway) BroadcastMessage(envelope *protocol.Envelope) {
	g.sessionManager.BroadcastToAll(envelope)
}

// SendToSession 向指定会话发送消息
func (g *SimpleGateway) SendToSession(sessionID string, envelope *protocol.Envelope) error {
	session, exists := g.sessionManager.GetSession(sessionID)
	if !exists {
		return ErrSessionNotFound
	}
	return session.SendEnvelope(envelope)
}

// sendWelcomeMessage 发送欢迎消息（内部方法）
func (g *SimpleGateway) sendWelcomeMessage(sess Session) {
	welcomeMsg := g.protocolManager.GetMessageFactory().CreateTextMessage("Welcome to Chat-Go!")
	_ = sess.SendEnvelope(welcomeMsg)
}
