package chat

import "time"

// SystemNoticeEvent 系统通知（例如管理员广播、服务状态提示等）
type SystemNoticeEvent struct {
	When    time.Time
	Level   string // info|warn|error
	Content string
}

func (e *SystemNoticeEvent) Type() EventType { return EventSystemNotice }
func (e *SystemNoticeEvent) Time() time.Time { return e.When }
