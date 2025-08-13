package transport

import (
	"encoding/json"
	"net"
	"testing"
	"time"
)

func TestFrameCodec_EncodeDecode_JSON(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	enc := NewFrameCodec(c1)
	dec := NewFrameCodec(c2)

	// Build a chat payload inside Envelope
	payload := mustJSON(ChatPayload{Content: "hello"})
	want := &Envelope{Type: "chat", From: "alice", Payload: payload, Ts: time.Now().UnixMilli()}
	go func() {
		if err := enc.Encode(want); err != nil {
			t.Errorf("encode error: %v", err)
		}
	}()

	var got Envelope
	if err := dec.Decode(&got, 1<<20); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.Type != want.Type || got.From != want.From {
		t.Fatalf("mismatch header: %+v vs %+v", got, want)
	}
	var gp, wp ChatPayload
	if err := json.Unmarshal(got.Payload, &gp); err != nil {
		t.Fatalf("unmarshal got payload: %v", err)
	}
	if err := json.Unmarshal(want.Payload, &wp); err != nil {
		t.Fatalf("unmarshal want payload: %v", err)
	}
	if gp.Content != wp.Content {
		t.Fatalf("payload mismatch: %+v vs %+v", gp, wp)
	}
}

func TestFrameCodec_MaxSize(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	enc := NewFrameCodec(c1)
	dec := NewFrameCodec(c2)

	// build large payload ~2MB
	big := make([]byte, 2<<20)
	for i := range big {
		if i%10 == 0 {
			big[i] = 'a'
		} else {
			big[i] = '0'
		}
	}
	msg := &Envelope{Type: "text", Payload: mustJSON(TextPayload{Text: string(big)})}
	go func() {
		_ = enc.Encode(msg)
	}()
	var got Envelope
	if err := dec.Decode(&got, 1<<20); err == nil {
		t.Fatalf("expected error due to max size, got nil")
	}
}
