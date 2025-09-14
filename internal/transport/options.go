package transport

import (
	"github.com/hongjun500/chat-go/internal/protocol"
	"time"
)

// Options configures transports (shared across TCP/WS where applicable)
type Options struct {
	OutBuffer    int           // client outgoing channel buffer size
	ReadTimeout  time.Duration // per-read deadline; 0 to disable
	WriteTimeout time.Duration // per-write deadline; 0 to disable
	MaxFrameSize int           // for framed transports (bytes), default 1MB

	// 新的协议管理器配置
	TCPProtocolManager *protocol.Manager // TCP 协议管理器
	WSProtocolManager  *protocol.Manager // WebSocket 协议管理器
}

// GetTCPProtocolManager 获取TCP协议管理器，如果未设置则根据编解码器类型创建
func (o *Options) GetTCPProtocolManager() *protocol.Manager {
	return o.TCPProtocolManager
}

// GetWSProtocolManager 获取WebSocket协议管理器，如果未设置则根据编解码器类型创建
func (o *Options) GetWSProtocolManager() *protocol.Manager {
	return o.WSProtocolManager
}
