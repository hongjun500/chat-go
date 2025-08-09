package chat

import "time"

// EventType 事件类型标识
type EventType string

const (
	EventUserJoined    EventType = "user.joined"
	EventUserLeave     EventType = "user.leave"
	EventMessageLocal  EventType = "message.local"
	EventMessageRemote EventType = "message.remote" // 来自远端节点
)

type Event interface {
	Type() EventType
	Time() time.Time
}
