package chat

import (
	"testing"
	"time"
)

func TestHubRegisterUnregister(t *testing.T) {
	hub := NewHub()
	c := NewClientWithBuffer("id1", 8)
	c.Name = "alice"
	hub.RegisterClient(c)

	names := hub.ListNames()
	if len(names) != 1 || names[0] != "alice" {
		t.Fatalf("expect one online 'alice', got %#v", names)
	}

	hub.UnregisterClient(c)
	if !c.IsClosed() {
		t.Fatalf("client should be closed after unregister")
	}
	if len(hub.ListNames()) != 0 {
		t.Fatalf("expect no online users")
	}
}

func TestSubscribeEmit(t *testing.T) {
	hub := NewHub()
	done := make(chan struct{}, 1)
	hub.Subscribe(EventHeartbeat, func(e Event) {
		if _, ok := e.(*HeartbeatEvent); ok {
			done <- struct{}{}
		}
	})
	hub.Emit(&HeartbeatEvent{When: time.Now(), FromID: "x"})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("heartbeat handler not invoked")
	}
}

func TestSendToAll(t *testing.T) {
	hub := NewHub()
	a := NewClientWithBuffer("a", 8)
	b := NewClientWithBuffer("b", 8)
	a.Name, b.Name = "alice", "bob"
	hub.RegisterClient(a)
	hub.RegisterClient(b)

	hub.SendToAll("hello")

	waitOne := func(c *Client) string {
		select {
		case s := <-c.Outgoing():
			return s
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting message for %s", c.Name)
			return ""
		}
	}
	sa := waitOne(a)
	sb := waitOne(b)
	if sa == "" || sb == "" {
		t.Fatalf("empty messages")
	}
}
