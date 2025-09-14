package protocol

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageFactory 负责创建各种类型的消息，统一消息创建逻辑
type MessageFactory struct {
	version string
}

// NewMessageFactory 创建消息工厂
func NewMessageFactory() *MessageFactory {
	return &MessageFactory{
		version: "1.0",
	}
}

// CreateTextMessage 创建文本消息
func (f *MessageFactory) CreateTextMessage(text string) *Envelope {
	return &Envelope{
		Version:  f.version,
		Type:     MsgText,
		Encoding: EncodingJSON,
		Mid:      uuid.New().String(),
		Ts:       time.Now().UnixMilli(),
		Data:     []byte(text),
	}
}

// CreateChatMessage 创建聊天消息
func (f *MessageFactory) CreateChatMessage(from, content string) *Envelope {
	return &Envelope{
		Version:  f.version,
		Type:     MsgText,
		Encoding: EncodingJSON,
		Mid:      uuid.New().String(),
		From:     from,
		Ts:       time.Now().UnixMilli(),
		Data:     []byte(content),
	}
}

// CreateSetNameMessage 创建设置昵称消息
func (f *MessageFactory) CreateSetNameMessage(name string) *Envelope {
	payload := SetNamePayload{Name: name}
	data, _ := json.Marshal(payload)

	return &Envelope{
		Version:  f.version,
		Type:     MsgText,
		Encoding: EncodingJSON,
		Mid:      uuid.New().String(),
		Ts:       time.Now().UnixMilli(),
		Data:     data,
	}
}

// CreateCommandMessage 创建命令消息
func (f *MessageFactory) CreateCommandMessage(command string) *Envelope {
	payload := CommandPayload{Raw: command}
	data, _ := json.Marshal(payload)

	return &Envelope{
		Version:  f.version,
		Type:     MsgCommand,
		Encoding: EncodingJSON,
		Mid:      uuid.New().String(),
		Ts:       time.Now().UnixMilli(),
		Data:     data,
	}
}

// CreateAckMessage 创建确认消息
func (f *MessageFactory) CreateAckMessage(status string, correlationID string) *Envelope {
	payload := AckPayload{Status: status}
	data, _ := json.Marshal(payload)

	return &Envelope{
		Version:     f.version,
		Type:        MsgAck,
		Encoding:    EncodingJSON,
		Mid:         uuid.New().String(),
		Correlation: correlationID,
		Ts:          time.Now().UnixMilli(),
		Data:        data,
	}
}

// CreatePingMessage 创建心跳ping消息
func (f *MessageFactory) CreatePingMessage(seq int64) *Envelope {
	payload := PingPayload{
		Seq:       seq,
		Timestamp: time.Now().UnixMilli(),
	}
	data, _ := json.Marshal(payload)

	return &Envelope{
		Version:  f.version,
		Type:     MsgPing,
		Encoding: EncodingJSON,
		Mid:      uuid.New().String(),
		Ts:       time.Now().UnixMilli(),
		Data:     data,
	}
}

// CreatePongMessage 创建心跳pong消息
func (f *MessageFactory) CreatePongMessage(seq int64, correlationID string) *Envelope {
	payload := PongPayload{
		Seq:       seq,
		Timestamp: time.Now().UnixMilli(),
	}
	data, _ := json.Marshal(payload)

	return &Envelope{
		Version:     f.version,
		Type:        MsgPong,
		Encoding:    EncodingJSON,
		Mid:         uuid.New().String(),
		Correlation: correlationID,
		Ts:          time.Now().UnixMilli(),
		Data:        data,
	}
}

// CreateDirectMessage 创建私聊消息
func (f *MessageFactory) CreateDirectMessage(from string, to []string, content string) *Envelope {
	payload := DirectPayload{
		To:      to,
		Content: content,
	}
	data, _ := json.Marshal(payload)

	return &Envelope{
		Version:  f.version,
		Type:     MsgText,
		Encoding: EncodingJSON,
		Mid:      uuid.New().String(),
		From:     from,
		Ts:       time.Now().UnixMilli(),
		Data:     data,
	}
}

// CreateFileMetaMessage 创建文件元数据消息
func (f *MessageFactory) CreateFileMetaMessage(from string, meta FileMetaPayload) *Envelope {
	data, _ := json.Marshal(meta)

	return &Envelope{
		Version:  f.version,
		Type:     MsgFileMeta,
		Encoding: EncodingJSON,
		Mid:      uuid.New().String(),
		From:     from,
		Ts:       time.Now().UnixMilli(),
		Data:     data,
	}
}

// CreateFileChunkMessage 创建文件分片消息
func (f *MessageFactory) CreateFileChunkMessage(from string, chunk FileChunkPayload) *Envelope {
	data, _ := json.Marshal(chunk)

	return &Envelope{
		Version:  f.version,
		Type:     MsgFileChunk,
		Encoding: EncodingJSON,
		Mid:      uuid.New().String(),
		From:     from,
		Ts:       time.Now().UnixMilli(),
		Data:     data,
	}
}
