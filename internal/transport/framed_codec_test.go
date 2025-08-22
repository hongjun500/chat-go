package transport

import (
	"encoding/json"
	"net"
	"testing"
)

func TestFramedMessageCodec_JSONEncoding(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	frameCodec1 := NewFrameCodec(c1)
	frameCodec2 := NewFrameCodec(c2)
	jsonCodec := &JSONCodec{}

	enc := NewFramedMessageCodec(frameCodec1, jsonCodec)
	dec := NewFramedMessageCodec(frameCodec2, jsonCodec)

	// Test message
	payload := mustJSON(ChatPayload{Content: "test message"})
	want := &Envelope{
		Type:    "chat",
		From:    "user1",
		Payload: payload,
	}

	// Encode and decode
	go func() {
		if err := enc.Encode(want); err != nil {
			t.Errorf("encode error: %v", err)
		}
	}()

	var got Envelope
	if err := dec.Decode(&got, 1<<20); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	// Verify
	if got.Type != want.Type || got.From != want.From {
		t.Fatalf("header mismatch: got %+v, want %+v", got, want)
	}

	var gotPayload, wantPayload ChatPayload
	if err := json.Unmarshal(got.Payload, &gotPayload); err != nil {
		t.Fatalf("unmarshal got payload: %v", err)
	}
	if err := json.Unmarshal(want.Payload, &wantPayload); err != nil {
		t.Fatalf("unmarshal want payload: %v", err)
	}
	if gotPayload.Content != wantPayload.Content {
		t.Fatalf("payload mismatch: got %+v, want %+v", gotPayload, wantPayload)
	}
}

func TestFramedMessageCodec_ProtobufEncoding(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	frameCodec1 := NewFrameCodec(c1)
	frameCodec2 := NewFrameCodec(c2)
	protoCodec := &ProtobufCodec{}

	enc := NewFramedMessageCodec(frameCodec1, protoCodec)
	dec := NewFramedMessageCodec(frameCodec2, protoCodec)

	// Test message
	want := &Envelope{
		Type: "test",
		From: "user1",
		Data: []byte("binary data"),
	}

	// Encode and decode
	go func() {
		if err := enc.Encode(want); err != nil {
			t.Errorf("encode error: %v", err)
		}
	}()

	var got Envelope
	if err := dec.Decode(&got, 1<<20); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	// Verify
	if got.Type != want.Type || got.From != want.From {
		t.Fatalf("header mismatch: got %+v, want %+v", got, want)
	}
	if string(got.Data) != string(want.Data) {
		t.Fatalf("data mismatch: got %s, want %s", string(got.Data), string(want.Data))
	}
}

func TestFramedMessageCodec_ContentType(t *testing.T) {
	frameCodec := NewFrameCodec(nil)
	
	jsonFramed := NewFramedMessageCodec(frameCodec, &JSONCodec{})
	if got := jsonFramed.ContentType(); got != ApplicationJson {
		t.Errorf("JSON content type: got %s, want %s", got, ApplicationJson)
	}

	protoFramed := NewFramedMessageCodec(frameCodec, &ProtobufCodec{})
	if got := protoFramed.ContentType(); got != ApplicationProtobuf {
		t.Errorf("Protobuf content type: got %s, want %s", got, ApplicationProtobuf)
	}
}