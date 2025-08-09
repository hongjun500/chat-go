package chat

import "time"

type Message struct {
	From      string
	Content   string
	Timestamp time.Time
}
