package transport

import (
	"context"
)

// Session 传输层统一的会话管理
type Session interface {
	ID() string
	RemoteAddr() string
	SendEnvelope(*Envelope) error // 发送消息（封装在 Envelope 结构体中）到客户端
	Close() error
}

// Gateway consumes high-level Envelope from any Transport
type Gateway interface {
	OnSessionOpen(sess Session)
	OnEnvelope(sess Session, msg *Envelope)
	OnSessionClose(sess Session, err error)
}

// Transport 统一的消息传输实现
type Transport interface {
	Name() string
	Start(ctx context.Context, addr string, gateway Gateway, opt Options) error
}
