package chat

import "time"

// UserEvent 表示用户加入 / 离开
type UserEvent struct {
	When time.Time
	User *Client
	Desc string // "joined" 或 "leave"
}

func (e *UserEvent) Type() EventType {
	if "leave" == e.Desc {
		return EventUserLeave
	}
	return EventUserJoined
}

func (e *UserEvent) Time() time.Time {
	return e.When
}
