package transport

import (
	"context"
)

const (
	Tcp       = "tcp"
	WebSocket = "websocket"
)

// Transport 统一的传输层接口
// 负责特定协议(TCP/WebSocket)的网络通信实现
type Transport interface {
	Name() string
	Start(ctx context.Context, addr string, gateway Gateway, opt Options) error
}
