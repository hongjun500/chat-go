package command

import (
	"testing"

	"github.com/hongjun500/chat-go/internal/chat"
)

func TestRegistryExecute_Basic(t *testing.T) {
	hub := chat.NewHub()
	reg := NewRegistry()
	// 注册一个简单命令
	err := reg.Register(&Command{
		Name: "echo",
		Help: "echo text",
		Handler: func(ctx *Context) error {
			ctx.Client.Send("ok:" + ctx.Raw)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("register err: %v", err)
	}

	c := chat.NewClientWithBuffer("c1", 4)
	ctx := &Context{Hub: hub, Client: c}

	handled, err := reg.Execute("/echo hi", ctx)
	if !handled || err != nil {
		t.Fatalf("execute failed: handled=%v err=%v", handled, err)
	}
	select {
	case s := <-c.Outgoing():
		if s == "" {
			t.Fatalf("empty resp")
		}
	default:
		t.Fatalf("no output queued")
	}
}
