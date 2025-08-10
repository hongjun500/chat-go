package redisstream

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Bus struct {
	cli    *redis.Client
	stream string
	group  string
}

type Message struct {
	Type string    `json:"type"`
	When time.Time `json:"when"`
	From string    `json:"from,omitempty"`
	To   string    `json:"to,omitempty"`
	Text string    `json:"text,omitempty"`
}

func New(addr string, db int, stream, group string) *Bus {
	cli := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	return &Bus{cli: cli, stream: stream, group: group}
}

func (b *Bus) EnsureGroup(ctx context.Context) error {
	// Create stream and group if not exist
	_ = b.cli.XGroupCreateMkStream(ctx, b.stream, b.group, "$").Err()
	return nil
}

func (b *Bus) Publish(ctx context.Context, m *Message) error {
	payload, _ := json.Marshal(m)
	return b.cli.XAdd(ctx, &redis.XAddArgs{Stream: b.stream, Values: map[string]any{"data": payload}}).Err()
}

type Handler func(ctx context.Context, m *Message) error

// Consume blocks and delivers messages to handler; call cancel to stop
func (b *Bus) Consume(ctx context.Context, consumer string, handler Handler) error {
	for {
		res, err := b.cli.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    b.group,
			Consumer: consumer,
			Streams:  []string{b.stream, ">"},
			Count:    100,
			Block:    5 * time.Second,
		}).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// transient errors: continue
			continue
		}
		for _, str := range res {
			for _, xmsg := range str.Messages {
				raw, _ := xmsg.Values["data"].(string)
				var m Message
				if err := json.Unmarshal([]byte(raw), &m); err == nil {
					_ = handler(ctx, &m)
				}
				// Acknowledge
				_ = b.cli.XAck(ctx, b.stream, b.group, xmsg.ID).Err()
			}
		}
	}
}
