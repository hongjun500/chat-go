package protocol

import "encoding/json"

// Encoding 表示消息负载的编码方式
type Encoding string

const (
	EncodingJSON     Encoding = "json"
	EncodingProtobuf Encoding = "protobuf"
	EncodingBinary   Encoding = "binary" // 保留扩展
)

// MessageType 表示系统支持的业务消息类型
type MessageType string

const (
	MsgText      MessageType = "text"
	MsgCommand   MessageType = "command"
	MsgFileMeta  MessageType = "file_meta"
	MsgFileChunk MessageType = "file_chunk"
	MsgAck       MessageType = "ack"
	MsgPing      MessageType = "ping"
	MsgPong      MessageType = "pong"
)

type Envelope struct {
	// ---- 协议元信息 ----
	Version  string      `json:"version"`  // 协议版本
	Type     MessageType `json:"type"`     // 消息类型
	Encoding Encoding    `json:"encoding"` // payload 编码方式

	// ---- 路由与可靠性 ----
	MessageID   string `json:"mid"`            // 消息唯一ID
	Correlation string `json:"correlation_id"` // 相关请求ID
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	Timestamp   int64  `json:"ts"` // 毫秒时间戳

	// ---- 负载 ----
	// 规则：
	// 1. 如果 Encoding=json → Payload 里就是 JSON RawMessage
	// 2. 如果 Encoding=protobuf → Data 里就是原始 Protobuf 字节（非 base64，直接走二进制帧）
	// 3. Payload/Data 二选一
	Payload json.RawMessage `json:"payload,omitempty"`
	Data    []byte          `json:"data,omitempty"`
}
