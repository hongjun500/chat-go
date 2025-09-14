package transport

import (
	"context"
	"github.com/hongjun500/chat-go/internal/protocol"
)

const (
	Tcp       = "tcp"
	WebSocket = "websocket"
)

// Session 传输层统一的会话管理接口
// 负责底层连接的生命周期管理和数据传输
type Session interface {
	ID() string
	RemoteAddr() string
	SendEnvelope(*protocol.Envelope) error // 发送消息到客户端
	Close() error
}

// Gateway 网关接口，处理传输层与上层业务的交互
// 负责处理会话事件和消息分发
type Gateway interface {
	OnSessionOpen(sess Session)
	OnEnvelope(sess Session, msg *protocol.Envelope)
	OnSessionClose(sess Session)
}

// Transport 统一的传输层接口
// 负责特定协议(TCP/WebSocket)的网络通信实现
type Transport interface {
	Name() string
	Start(ctx context.Context, addr string, gateway Gateway, opt Options) error
}

// MessageCodecProvider 提供消息编解码器的接口
// 用于传输层获取协议层的编解码能力，保持层间解耦
type MessageCodecProvider interface {
	GetCodec() protocol.MessageCodec
}
