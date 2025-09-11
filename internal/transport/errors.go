package transport

import "errors"

// 传输层错误定义
var (
	ErrSessionClosed = errors.New("session is closed")
	ErrInvalidCodec  = errors.New("invalid codec type")
	ErrFrameTooLarge = errors.New("frame too large")
)