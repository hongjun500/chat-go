package chat

import "time"

// HeartbeatEvent 心跳事件，用于在线状态保活、丢弃统计或跨节点同步
type HeartbeatEvent struct {
	When   time.Time
	FromID string
	Detail string // 可选：延迟、客户端信息等
}

func (e *HeartbeatEvent) Type() EventType { return EventHeartbeat }
func (e *HeartbeatEvent) Time() time.Time { return e.When }
