package transport

import "time"

// Options configures transports (shared across TCP/WS where applicable)
type Options struct {
	OutBuffer    int           // client outgoing channel buffer size
	ReadTimeout  time.Duration // per-read deadline; 0 to disable
	WriteTimeout time.Duration // per-write deadline; 0 to disable
	MaxFrameSize int           // for framed transports (bytes), default 1MB
}
