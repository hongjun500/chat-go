package protocol

import "encoding/json"

// Encoding 表示消息负载的编码方式
type Encoding string

const (
	EncodingJSON     Encoding = Json
	EncodingProtobuf Encoding = Protobuf
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
	Mid         string `json:"mid"`            // 消息唯一ID
	Correlation string `json:"correlation_id"` // 相关请求ID
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	Ts          int64  `json:"ts"` // 毫秒时间戳

	// ---- 负载 ----
	// 规则：
	// 1. 如果 Encoding=json → Payload 里就是 JSON RawMessage
	// 2. 如果 Encoding=protobuf → Data 里就是原始 Protobuf 字节（非 base64，直接走二进制帧）
	// 3. Payload/Data 二选一
	Payload json.RawMessage `json:"payload,omitempty"`
	Data    []byte          `json:"data,omitempty"`
}

// TextPayload 纯文本消息负载
type TextPayload struct {
	Text string `json:"text"`
}

// SetNamePayload 设置昵称消息负载
type SetNamePayload struct {
	Name string `json:"name"`
}

// ChatPayload 聊天消息负载
type ChatPayload struct {
	Content string `json:"content"`
}

// CommandPayload 命令消息负载
type CommandPayload struct {
	Raw string `json:"raw"`
}

// AckPayload 确认消息负载
type AckPayload struct {
	Status string `json:"status"`
}

// DirectPayload 私聊消息负载
type DirectPayload struct {
	To      []string `json:"to"`
	Content string   `json:"content"`
}

// PingPayload 心跳 ping 消息负载
type PingPayload struct {
	Seq       int64 `json:"seq"`
	Timestamp int64 `json:"timestamp"`
}

// PongPayload 心跳 pong 消息负载
type PongPayload struct {
	Seq       int64 `json:"seq"`
	Timestamp int64 `json:"timestamp"`
}

// FileMetaPayload 文件元数据消息负载
type FileMetaPayload struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	Checksum string `json:"checksum"`
}

// FileChunkPayload 文件分片消息负载
type FileChunkPayload struct {
	FileID   string `json:"file_id"`
	ChunkID  int    `json:"chunk_id"`
	Data     []byte `json:"data"`
	IsLast   bool   `json:"is_last"`
	Checksum string `json:"checksum"`
}
