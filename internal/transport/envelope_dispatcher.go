package transport

import (
	"fmt"
	"github.com/hongjun500/chat-go/internal/protocol"
)

// EnvelopeDispatcher 信封分发器，使用新的消息路由器
type EnvelopeDispatcher struct {
	protocol *protocol.Protocol
	router   *protocol.MessageRouter
}

// NewEnvelopeDispatcher 创建信封分发器
func NewEnvelopeDispatcher(codecType int) *EnvelopeDispatcher {
	d := &EnvelopeDispatcher{
		protocol: protocol.NewProtocol(codecType),
		router:   protocol.NewMessageRouter(),
	}
	
	// 注册默认的文本处理器
	d.router.RegisterHandler(protocol.MsgText, d.textHandler)
	
	return d
}

// Welcome 发送欢迎消息
func (d *EnvelopeDispatcher) Welcome(sess Session) {
	envelope := d.protocol.Welcome("请输入昵称并回车：")
	_ = sess.SendEnvelope(envelope)
}

// Dispatch 分发消息
func (d *EnvelopeDispatcher) Dispatch(sess Session, e *protocol.Envelope) error {
	return d.router.Dispatch(e)
}

// RegisterHandler 注册消息处理器
func (d *EnvelopeDispatcher) RegisterHandler(msgType protocol.MessageType, handler protocol.MessageHandler) {
	d.router.RegisterHandler(msgType, handler)
}

// textHandler 文本消息处理器（示例实现）
func (d *EnvelopeDispatcher) textHandler(env *protocol.Envelope) error {
	// 这里是简化的实现，实际应该由业务层处理
	// 比如设置昵称、聊天等逻辑
	fmt.Printf("Received text message: %s\n", string(env.Data))
	return nil
}
