package protocol

import (
	"fmt"
	"sync"
)

// MessageHandler 处理消息的函数类型
type MessageHandler func(env *Envelope) error

// MessageRouter 专门负责消息路由和分发
type MessageRouter struct {
	mu             sync.RWMutex
	handlers       map[MessageType]MessageHandler
	defaultHandler MessageHandler
}

// NewMessageRouter 创建新的消息路由器
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		handlers: make(map[MessageType]MessageHandler),
	}
}

// RegisterHandler 注册消息类型处理函数
func (r *MessageRouter) RegisterHandler(msgType MessageType, handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[msgType] = handler
}

// SetDefaultHandler 设置默认处理函数
func (r *MessageRouter) SetDefaultHandler(handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultHandler = handler
}

// Dispatch 分发消息到对应处理函数
func (r *MessageRouter) Dispatch(env *Envelope) error {
	r.mu.RLock()
	handler, ok := r.handlers[env.Type]
	r.mu.RUnlock()
	
	if ok {
		return handler(env)
	}
	
	r.mu.RLock()
	defaultHandler := r.defaultHandler
	r.mu.RUnlock()
	
	if defaultHandler != nil {
		return defaultHandler(env)
	}
	
	return fmt.Errorf("no handler registered for message type: %s", env.Type)
}

// GetRegisteredTypes 获取已注册的消息类型列表
func (r *MessageRouter) GetRegisteredTypes() []MessageType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	types := make([]MessageType, 0, len(r.handlers))
	for msgType := range r.handlers {
		types = append(types, msgType)
	}
	return types
}