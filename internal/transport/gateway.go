package transport

import (
	"encoding/json"
	"sync"

	"github.com/hongjun500/chat-go/internal/protocol"
	"github.com/hongjun500/chat-go/pkg/logger"
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

	initOnce sync.Once
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
	logger.L().Sugar().Infow("OnSessionOpen", "SessionId", sc.Id, "addr", sc.RemoteAddr)
	// 使用已有的 SessionContext 注册到会话管理器，保持上下文对象一致性
	g.sessionManager.AddContext(sc)

	// todo 放到 网关的 dispatcher
	// 初始化内置处理器（只执行一次）
	g.initOnce.Do(func() {
		// 处理 ping -> 返回 pong
		g.disp.Register(string(protocol.MsgPing), func(ctx *SessionContext, msg *protocol.Envelope) {
			// 试图解析 ping payload
			var p protocol.PingPayload
			_ = json.Unmarshal(msg.Data, &p) // 若解析失败，仍然返回 pong（seq 为 0）
			// 创建 pong（使用原 message id 作为 correlation）
			pong := protocol.NewMessageFactory().CreatePongMessage(p.Seq, msg.Mid)
			if err := ctx.Send(pong); err != nil {
				logger.L().Sugar().Warnw("send_pong_failed", "session", ctx.Id, "err", err)
			}
		})

		// 可在此注册更多基础处理器（如 command、ack 等），保持简洁最小化实现
	})

	// 发送欢迎消息（非阻塞，记录错误）
	welcome := protocol.NewMessageFactory().CreateTextMessage("Welcome to Chat-Go!")
	if err := sc.Send(welcome); err != nil {
		logger.L().Sugar().Warnw("send_welcome_failed", "session", sc.Id, "err", err)
	}
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
