package transport

import (
	"github.com/hongjun500/chat-go/internal/protocol"
	"log"
)

type HandlerFunc func(ctx *SessionContext, msg *protocol.Envelope)

type dispatcher struct {
	handlers map[string]HandlerFunc
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		handlers: make(map[string]HandlerFunc),
	}
}

// Register 注册处理器
func (d *dispatcher) Register(msgType string, handler HandlerFunc) {
	d.handlers[msgType] = handler
}

// Dispatch 分发消息
func (d *dispatcher) Dispatch(ctx *SessionContext, msg *protocol.Envelope) {
	if handler, ok := d.handlers[string(msg.Type)]; ok {
		handler(ctx, msg)
	} else {
		log.Printf("no handler for message type: %s", msg.Type)
	}
}
