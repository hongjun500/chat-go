package protocol

import (
	"fmt"
	"strings"
	"sync"
)

type MessageHandler func(env *Envelope) error

type Protocol struct {
	mu             sync.RWMutex
	handlers       map[MessageType]MessageHandler
	defaultHandler MessageHandler
	Codec          MessageCodec
	config         CodecConfig
}

var DefaultProtocol *Protocol

func init() {
	// 注册内置的 text 消息处理函数
	DefaultProtocol = NewProtocol(CodecJson)
	DefaultProtocol.SetDefaultHandler(DefaultProtocol.textHandler)
}

func NewProtocol(codecType int) *Protocol {
	config := NewDefaultCodecConfig(codecType)
	factory, exists := CodecFactories[codecType]
	if !exists {
		// 回退到JSON编码器作为默认
		factory = CodecFactories[CodecJson]
		config = NewDefaultCodecConfig(CodecJson)
	}
	codec := factory()
	return &Protocol{
		handlers: make(map[MessageType]MessageHandler),
		Codec:    codec,
		config:   config,
	}
}

// NewProtocolWithConfig 使用配置创建协议
func NewProtocolWithConfig(config CodecConfig) (*Protocol, error) {
	codecType := config.GetDefaultCodec()
	codec, err := NewCodec(codecType)
	if err != nil {
		return nil, fmt.Errorf("failed to create codec: %w", err)
	}
	return &Protocol{
		handlers: make(map[MessageType]MessageHandler),
		Codec:    codec,
		config:   config,
	}, nil
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
		Data: []byte(text),
	}
	return e, nil
}

func (d *Protocol) textHandler(env *Envelope) error {

	return nil
}

// SetCodec 动态设置编码器
func (d *Protocol) SetCodec(codecType int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if !d.config.IsCodecSupported(codecType) {
		return fmt.Errorf("unsupported codec type: %d", codecType)
	}
	
	codec, err := NewCodec(codecType)
	if err != nil {
		return fmt.Errorf("failed to create codec: %w", err)
	}
	
	d.Codec = codec
	return d.config.SetDefaultCodec(codecType)
}

// SetCodecByName 根据名称动态设置编码器
func (d *Protocol) SetCodecByName(name string) error {
	codecType, err := d.config.GetCodecByName(name)
	if err != nil {
		return fmt.Errorf("failed to get codec by name: %w", err)
	}
	return d.SetCodec(codecType)
}

// GetCurrentCodec 获取当前编码器信息
func (d *Protocol) GetCurrentCodec() (string, int) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	codecType := d.config.GetDefaultCodec()
	codecName := CodecTypeMapping[codecType]
	return codecName, codecType
}

// GetConfig 获取编码器配置
func (d *Protocol) GetConfig() CodecConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.config
}
