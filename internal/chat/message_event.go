package chat

import "time"

// MessageEvent 表示一条聊天消息（包含系统 / 用户消息）
type MessageEvent struct {
	When    time.Time
	From    string
	Content string
	Local   bool // 本地生成还是远端同步
}

func (e *MessageEvent) Type() EventType {
	if e.Local {
		return EventMessageLocal
	}
	return EventMessageRemote
}

func (e *MessageEvent) Time() time.Time { return e.When }
