package transport

import (
	"fmt"
)

// 传输层错误定义
var (
	ErrSessionClosed = NewTpError(1001, "Session is closed", "")
	ErrInvalidCodec  = NewTpError(1002, "Invalid codec type", "")
	ErrFrameTooLarge = NewTpError(1003, "Frame too large", "")
)

type tpError struct {
	code    int
	msg     string
	context string
}

func (e *tpError) Error() string {
	if e.context != "" {
		return fmt.Sprintf("Error %d: %s (context: %s)", e.code, e.msg, e.context)
	}
	return fmt.Sprintf("Error %d: %s", e.code, e.msg)
}

func NewTpError(code int, message string, context string) *tpError {
	return &tpError{
		code:    code,
		msg:     message,
		context: context,
	}
}
