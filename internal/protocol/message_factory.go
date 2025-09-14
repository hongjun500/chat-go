package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// MessageFactory 消息工厂，负责创建和编解码消息
type MessageFactory struct {
	defaultCodec      MessageCodec
	versionController *VersionController
	config           CodecConfig
}

// NewMessageFactory 创建消息工厂
func NewMessageFactory(defaultCodec MessageCodec, config CodecConfig) *MessageFactory {
	return &MessageFactory{
		defaultCodec: defaultCodec,
		config:      config,
	}
}

// NewMessageFactoryWithVersionControl 创建支持版本控制的消息工厂
func NewMessageFactoryWithVersionControl(defaultCodec MessageCodec, config CodecConfig, versionController *VersionController) *MessageFactory {
	return &MessageFactory{
		defaultCodec:      defaultCodec,
		config:           config,
		versionController: versionController,
	}
}

// CreateTextMessage 创建文本消息
func (mf *MessageFactory) CreateTextMessage(text, from, to string) (*Envelope, error) {
	if text == "" {
		return nil, fmt.Errorf("MessageFactory.CreateTextMessage: text cannot be empty")
	}
	
	payload := TextPayload{Text: text}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("MessageFactory.CreateTextMessage: failed to marshal payload: %w", err)
	}
	
	return &Envelope{
		Version:  "1.0",
		Type:     MsgText,
		Encoding: EncodingJSON,
		Mid:      generateMessageID(),
		From:     from,
		To:       to,
		Ts:       time.Now().UnixMilli(),
		Data:     data,
	}, nil
}

// CreateChatMessage 创建聊天消息
func (mf *MessageFactory) CreateChatMessage(content, from string) (*Envelope, error) {
	if content == "" {
		return nil, fmt.Errorf("MessageFactory.CreateChatMessage: content cannot be empty")
	}
	
	payload := ChatPayload{Content: content}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("MessageFactory.CreateChatMessage: failed to marshal payload: %w", err)
	}
	
	return &Envelope{
		Version:  "1.0",
		Type:     "chat",
		Encoding: EncodingJSON,
		Mid:      generateMessageID(),
		From:     from,
		Ts:       time.Now().UnixMilli(),
		Data:     data,
	}, nil
}

// CreateAckMessage 创建确认消息
func (mf *MessageFactory) CreateAckMessage(status, correlationID string) (*Envelope, error) {
	if status == "" {
		return nil, fmt.Errorf("MessageFactory.CreateAckMessage: status cannot be empty")
	}
	
	payload := AckPayload{Status: status}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("MessageFactory.CreateAckMessage: failed to marshal payload: %w", err)
	}
	
	return &Envelope{
		Version:     "1.0",
		Type:        MsgAck,
		Encoding:    EncodingJSON,
		Mid:         generateMessageID(),
		Correlation: correlationID,
		Ts:          time.Now().UnixMilli(),
		Data:        data,
	}, nil
}

// CreateCommandMessage 创建命令消息
func (mf *MessageFactory) CreateCommandMessage(command, from string) (*Envelope, error) {
	if command == "" {
		return nil, fmt.Errorf("MessageFactory.CreateCommandMessage: command cannot be empty")
	}
	
	payload := CommandPayload{Raw: command}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("MessageFactory.CreateCommandMessage: failed to marshal payload: %w", err)
	}
	
	return &Envelope{
		Version:  "1.0",
		Type:     MsgCommand,
		Encoding: EncodingJSON,
		Mid:      generateMessageID(),
		From:     from,
		Ts:       time.Now().UnixMilli(),
		Data:     data,
	}, nil
}

// EncodeMessage 编码消息
func (mf *MessageFactory) EncodeMessage(env *Envelope) ([]byte, error) {
	if env == nil {
		return nil, fmt.Errorf("MessageFactory.EncodeMessage: envelope is nil")
	}
	
	codec, err := mf.getCodecForMessage(env)
	if err != nil {
		return nil, fmt.Errorf("MessageFactory.EncodeMessage: failed to get codec: %w", err)
	}
	
	var buf bytes.Buffer
	if err := codec.Encode(&buf, env); err != nil {
		return nil, fmt.Errorf("MessageFactory.EncodeMessage: encoding failed with %s codec: %w", codec.Name(), err)
	}
	
	return buf.Bytes(), nil
}

// DecodeMessage 解码消息
func (mf *MessageFactory) DecodeMessage(data []byte, maxSize int) (*Envelope, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("MessageFactory.DecodeMessage: data is empty")
	}
	
	if maxSize > 0 && len(data) > maxSize {
		return nil, fmt.Errorf("MessageFactory.DecodeMessage: data size %d exceeds max size %d", len(data), maxSize)
	}
	
	var env Envelope
	reader := bytes.NewReader(data)
	
	// 首先尝试使用默认编码器
	if err := mf.defaultCodec.Decode(reader, &env, maxSize); err == nil {
		return &env, nil
	}
	
	// 如果默认编码器失败，尝试其他编码器
	for codecType := range CodecFactories {
		codec, err := NewCodec(codecType)
		if err != nil {
			continue
		}
		
		// 重置reader
		reader.Reset(data)
		env = Envelope{} // 重置envelope
		
		if err := codec.Decode(reader, &env, maxSize); err == nil {
			return &env, nil
		}
	}
	
	return nil, fmt.Errorf("MessageFactory.DecodeMessage: failed to decode with any available codec")
}

// DecodeMessageWithCodec 使用指定编码器解码消息
func (mf *MessageFactory) DecodeMessageWithCodec(data []byte, codecType int, maxSize int) (*Envelope, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("MessageFactory.DecodeMessageWithCodec: data is empty")
	}
	
	codec, err := NewCodec(codecType)
	if err != nil {
		return nil, fmt.Errorf("MessageFactory.DecodeMessageWithCodec: failed to create codec: %w", err)
	}
	
	var env Envelope
	reader := bytes.NewReader(data)
	
	if err := codec.Decode(reader, &env, maxSize); err != nil {
		return nil, fmt.Errorf("MessageFactory.DecodeMessageWithCodec: decoding failed with %s codec: %w", codec.Name(), err)
	}
	
	return &env, nil
}

// getCodecForMessage 获取消息对应的编码器
func (mf *MessageFactory) getCodecForMessage(env *Envelope) (MessageCodec, error) {
	// 如果有版本控制器，使用版本特定的编码器
	if mf.versionController != nil {
		codec, err := mf.versionController.GetCodecForMessage(env)
		if err == nil {
			return codec, nil
		}
	}
	
	// 回退到默认编码器
	return mf.defaultCodec, nil
}

// SetDefaultCodec 设置默认编码器
func (mf *MessageFactory) SetDefaultCodec(codec MessageCodec) {
	mf.defaultCodec = codec
}

// GetDefaultCodec 获取默认编码器
func (mf *MessageFactory) GetDefaultCodec() MessageCodec {
	return mf.defaultCodec
}

// generateMessageID 生成消息ID (简单实现)
func generateMessageID() string {
	return fmt.Sprintf("msg-%d", time.Now().UnixNano())
}