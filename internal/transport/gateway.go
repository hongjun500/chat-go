package transport

import (
	"github.com/hongjun500/chat-go/internal/protocol"
)

type handlerFunc func(Session, *protocol.Envelope)

// Gateway 网关接口，*SessionContext 的交互
// 负责处理会话事件和消息分发
type Gateway interface {
	OnSessionOpen(sc *SessionContext)
	OnEnvelope(sc *SessionContext, msg *protocol.Envelope)
	OnSessionClose(sc *SessionContext)
}

// SimpleGateway 简单的网关实现，专注于消息转发和会话管理
// 作为传输层与业务层的桥梁
type SimpleGateway struct {
	sessionManager *SessionManager
	disp           *dispatcher
}

// NewSimpleGateway 创建简单网关
func NewSimpleGateway() *SimpleGateway {
	return &SimpleGateway{
		sessionManager: NewSessionManager(),
		disp:           newDispatcher(),
	}
}

// OnSessionOpen 会话开启事件
func (g *SimpleGateway) OnSessionOpen(sc *SessionContext) {
	g.sessionManager.Add(sc.sess)
	// todo 发送欢迎消息或进行会话初始化
	// sc.Send()
}

// OnEnvelope 处理收到的消息
func (g *SimpleGateway) OnEnvelope(sc *SessionContext, msg *protocol.Envelope) {
	// todo 回退到 tp 层的全局处理器
	g.disp.Dispatch(sc, msg)
}

// OnSessionClose 会话关闭事件
func (g *SimpleGateway) OnSessionClose(sc *SessionContext) {
	defer g.sessionManager.Remove(sc.Id)
	_ = sc.Close()
}

// GetSessionManager 获取会话管理器
func (g *SimpleGateway) GetSessionManager() *SessionManager {
	return g.sessionManager
}

// GetSession 获取指定会话
func (g *SimpleGateway) GetSession(sessionID string) (*SessionContext, bool) {
	return g.sessionManager.Get(sessionID)
}

/*// BroadcastMessage 向所有会话广播消息
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
func (g *SimpleGateway) sendWelcomeMessage(sc Session) {
	session, b := g.sessionManager.GetSession(sc.ID())

	welcomeMsg := g.protocolManager.GetMessageFactory().CreateTextMessage("Welcome to Chat-Go!")
	_ = sc.SendEnvelope(welcomeMsg)
}*/
