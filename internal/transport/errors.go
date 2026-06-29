package transport

import (
	"errors"
	"fmt"
)

// tp errs
var (
	ErrSessionClosed   = errors.New("session is closed")
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidFrame    = errors.New("invalid frame format")
	ErrFrameTooLarge   = errors.New("frame size exceeds maximum allowed")
	ErrConnectionLost  = errors.New("connection lost")

	ErrSessionContextClosed = newTpError(1001, "Session context is closed", "")
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

func newTpError(code int, message string, context string) *tpError {
	return &tpError{
		code:    code,
		msg:     message,
		context: context,
	}
}
