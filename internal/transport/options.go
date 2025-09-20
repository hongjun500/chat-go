package transport

import (
	"github.com/hongjun500/chat-go/internal/protocol"
	"time"
)

// Options configures transports (shared across TCP/WS where applicable)
type Options struct {
	OutBuffer         int           // client outgoing channel buffer size
	ReadTimeout       time.Duration // 读取超时
	WriteTimeout      time.Duration // 写入超时
	HeartbeatInterval time.Duration // 心跳间隔，（服务端检测间隔
	HeartbeatTimeout  time.Duration // 心跳超时（客户端允许多长时间不发心跳）
	MaxFrameSize      int           // for framed transports (bytes), default 1MB

	// 新的协议管理器配置
	TCPProtocolManager *protocol.Manager // TCP 协议管理器
	WSProtocolManager  *protocol.Manager // WebSocket 协议管理器
}

// GetTCPProtocolManager TCP 协议管理器
func (o *Options) GetTCPProtocolManager() *protocol.Manager {
	return o.TCPProtocolManager
}

// GetWSProtocolManager WebSocket 协议管理器
func (o *Options) GetWSProtocolManager() *protocol.Manager {
	return o.WSProtocolManager
}
