package transport

import (
	"github.com/hongjun500/chat-go/internal/protocol"
)

// Gateway 接口（保持不变）
// Gateway consumes high-level Envelope from any Transport
// 注意：这里不应该包含具体的业务逻辑，只做消息转发

// SimpleGateway 简单的网关实现，专注于消息转发
type SimpleGateway struct {
	sessionManager *SessionManager
	dispatcher     *EnvelopeDispatcher
	handlers       map[string]func(Session, *protocol.Envelope)
}

// NewSimpleGateway 创建简单网关
func NewSimpleGateway(codecType int) *SimpleGateway {
	return &SimpleGateway{
		sessionManager: NewSessionManager(),
		dispatcher:     NewEnvelopeDispatcher(codecType),
		handlers:       make(map[string]func(Session, *protocol.Envelope)),
	}
}

// RegisterHandler 注册会话级别的处理器
func (g *SimpleGateway) RegisterHandler(msgType string, handler func(Session, *protocol.Envelope)) {
	g.handlers[msgType] = handler
}

// OnSessionOpen 会话开启事件
func (g *SimpleGateway) OnSessionOpen(sess Session) {
	g.sessionManager.AddSession(sess)
	g.dispatcher.Welcome(sess)
}

// OnEnvelope 处理收到的消息
func (g *SimpleGateway) OnEnvelope(sess Session, msg *protocol.Envelope) {
	// 首先尝试会话级别的处理器
	if handler, exists := g.handlers[string(msg.Type)]; exists {
		handler(sess, msg)
		return
	}
	
	// 回退到默认分发器
	_ = g.dispatcher.Dispatch(sess, msg)
}

// OnSessionClose 会话关闭事件
func (g *SimpleGateway) OnSessionClose(sess Session, err error) {
	g.sessionManager.RemoveSession(sess.ID())
}

// GetSessionManager 获取会话管理器
func (g *SimpleGateway) GetSessionManager() *SessionManager {
	return g.sessionManager
}
