package protocol

import (
	"fmt"
	"go.uber.org/zap/buffer"
	"strings"
	"sync"
)

type MessageHandler func(env *Envelope) error

type Protocol struct {
	mu             sync.RWMutex
	handlers       map[MessageType]MessageHandler
	defaultHandler MessageHandler
	codec          MessageCodec
}

var DefaultProtocol *Protocol

func init() {
	// 注册内置的 text 消息处理函数
	DefaultProtocol := NewProtocol(CodecJson)
	DefaultProtocol.SetDefaultHandler(DefaultProtocol.textHandler)
}

func NewProtocol(codecType int) *Protocol {
	return &Protocol{
		handlers: make(map[MessageType]MessageHandler),
		codec:    CodecFactories[codecType](),
	}
}

func (d *Protocol) RegisterHandler(msgType MessageType, handler MessageHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[msgType] = handler
}

// SetDefaultHandler 设置默认处理函数
func (d *Protocol) SetDefaultHandler(handler MessageHandler) {
	d.defaultHandler = handler
}

// Dispatch 分发消息到对应处理函数
func (d *Protocol) Dispatch(env *Envelope) error {
	d.mu.RLock()
	handler, ok := d.handlers[env.Type]
	d.mu.RUnlock()
	if ok {
		return handler(env)
	}
	if d.defaultHandler != nil {
		return d.defaultHandler(env)
	}
	return fmt.Errorf("no handler registered for message type: %s", env.Type)
}

// Welcome 发送欢迎消息
func (d *Protocol) Welcome(text string) (*Envelope, error) {
	text = strings.TrimSpace(text)
	e := &Envelope{
		Type: MsgText,
		Ts:   0,
		Data: nil,
	}
	var buff buffer.Buffer
	if err := d.codec.Encode(&buff, e); err != nil {
		return nil, err
	}
	return e, nil
}

func (d *Protocol) textHandler(env *Envelope) error {

	return nil
}
