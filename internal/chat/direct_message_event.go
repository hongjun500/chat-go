package chat

import "time"

type DirectMessageEvent struct {
	When    time.Time
	From    string
	To      string
	Content string
}

func (e *DirectMessageEvent) Type() EventType { return EventMessageDirect }
func (e *DirectMessageEvent) Time() time.Time { return e.When }
