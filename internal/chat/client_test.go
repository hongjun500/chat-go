package chat

import (
	"testing"
	"time"
)

func TestClientClose(t *testing.T) {
	c := NewClientWithBuffer("id1", 2)
	c.Send("hello")

	if c.IsClosed() {
		t.Fatalf("client should be open before Close")
	}

	c.Close()
	if !c.IsClosed() {
		t.Fatalf("client should be closed after Close")
	}

	// Outgoing should be closed: second receive should return ok=false quickly
	select {
	case msg, ok := <-c.Outgoing():
		// first receive may still drain remaining buffered message
		_ = msg
		if !ok {
			// already closed and drained — acceptable
			return
		}
		// Check next read returns closed
		select {
		case _, ok2 := <-c.Outgoing():
			if ok2 {
				t.Fatalf("outgoing channel should be closed after draining buffered messages")
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timeout waiting for closed outgoing channel")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting to read from outgoing after Close")
	}
}

func TestClientSendDropWhenBufferFull(t *testing.T) {
	// buffer size = 1, send two messages; second should be dropped
	c := NewClientWithBuffer("id2", 1)
	c.Send("a")
	c.Send("b") // should be dropped due to full buffer

	// Drain first
	var first string
	select {
	case first = <-c.Outgoing():
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting first message")
	}
	if first != "a" {
		t.Fatalf("expected first message 'a', got %q", first)
	}

	// There should be no second message available (dropped)
	select {
	case m := <-c.Outgoing():
		// If any, this means drop policy failed or buffer > 1
		t.Fatalf("expected no second message, got %q", m)
	default:
		// ok, no more messages
	}
}
