package chat

import "time"

// EventType 事件类型标识
type EventType string

const (
	EventUserJoined    EventType = "user.joined"
	EventUserLeave     EventType = "user.leave"
	EventMessageLocal  EventType = "message.local"
	EventMessageRemote EventType = "message.remote" // 来自远端节点
	// 新增：系统通知、文件传输、心跳
	EventSystemNotice EventType = "system.notice"
	EventFileTransfer EventType = "file.transfer"
	EventHeartbeat    EventType = "heartbeat"
)

type Event interface {
	Type() EventType
	Time() time.Time
}
