package transport

import "time"

// Options configures transports (shared across TCP/WS where applicable)
type Options struct {
	OutBuffer    int           // client outgoing channel buffer size
	ReadTimeout  time.Duration // per-read deadline; 0 to disable
	WriteTimeout time.Duration // per-write deadline; 0 to disable
	MaxFrameSize int           // for framed transports (bytes), default 1MB
	TCPCodec     int           // TCP 编解码器类型 (0:JSON, 1:Protobuf)
	WSCodec      int           // WebSocket 编解码器类型 (0:JSON, 1:Protobuf)
}
