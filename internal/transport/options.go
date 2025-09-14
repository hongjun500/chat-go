package transport

import (
	"time"
	"github.com/hongjun500/chat-go/internal/protocol"
)

// Options configures transports (shared across TCP/WS where applicable)
type Options struct {
	OutBuffer     int           // client outgoing channel buffer size
	ReadTimeout   time.Duration // per-read deadline; 0 to disable
	WriteTimeout  time.Duration // per-write deadline; 0 to disable
	MaxFrameSize  int           // for framed transports (bytes), default 1MB
	
	// 新的协议管理器配置
	TCPProtocolManager *protocol.ProtocolManager // TCP 协议管理器
	WSProtocolManager  *protocol.ProtocolManager // WebSocket 协议管理器
	
	// 向后兼容的编解码器类型（已废弃）
	TCPCodec int // Deprecated: 使用 TCPProtocolManager 替代
	WSCodec  int // Deprecated: 使用 WSProtocolManager 替代
}

// GetTCPProtocolManager 获取TCP协议管理器，如果未设置则根据编解码器类型创建
func (o *Options) GetTCPProtocolManager() *protocol.ProtocolManager {
	if o.TCPProtocolManager != nil {
		return o.TCPProtocolManager
	}
	// 向后兼容：根据编解码器类型创建协议管理器
	return protocol.NewProtocolManager(o.TCPCodec)
}

// GetWSProtocolManager 获取WebSocket协议管理器，如果未设置则根据编解码器类型创建
func (o *Options) GetWSProtocolManager() *protocol.ProtocolManager {
	if o.WSProtocolManager != nil {
		return o.WSProtocolManager
	}
	// 向后兼容：根据编解码器类型创建协议管理器
	return protocol.NewProtocolManager(o.WSCodec)
}
