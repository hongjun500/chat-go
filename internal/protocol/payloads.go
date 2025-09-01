package protocol

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
	FileID    string `json:"file_id"`
	ChunkID   int    `json:"chunk_id"`
	Data      []byte `json:"data"`
	IsLast    bool   `json:"is_last"`
	Checksum  string `json:"checksum"`
}