package chat

import "time"

// FileTransferEvent 文件传输事件（仅元数据，实际数据应走独立通道/存储）
type FileTransferEvent struct {
	When       time.Time
	From       string
	To         string // 为空表示群发
	FileName   string
	SizeBytes  int64
	MimeType   string
	StorageKey string // 存储位置（如对象存储 Key/URL 等）
}

func (e *FileTransferEvent) Type() EventType { return EventFileTransfer }
func (e *FileTransferEvent) Time() time.Time { return e.When }
