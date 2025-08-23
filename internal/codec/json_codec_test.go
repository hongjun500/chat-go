package codec

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/hongjun500/chat-go/internal/protocol"
	"io"
	"strings"
	"testing"
)

func TestJSONCodec_ContentType(t *testing.T) {
	codec := JSONCodec{}
	if got := codec.ContentType(); got != ApplicationJson {
		t.Errorf("ContentType() = %v, want %v", got, ApplicationJson)
	}
}

func TestJSONCodec_EncodeDecode(t *testing.T) {
	codec := JSONCodec{}
	payload, _ := json.Marshal(protocol.ChatPayload{Content: "hello world"})
	orig := &protocol.Envelope{
		Type:    "chat",
		Payload: payload,
		From:    "user1",
		To:      []string{"user2"},
	}

	var buf bytes.Buffer
	if err := codec.Encode(&buf, orig); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded := &protocol.Envelope{}
	if err := codec.Decode(&buf, decoded, 0); err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.Type != orig.Type || decoded.From != orig.From {
		t.Errorf("Decoded header mismatch. Got %+v, want %+v", decoded, orig)
	}
	var dp, op protocol.ChatPayload
	if err := json.Unmarshal(decoded.Payload, &dp); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if err := json.Unmarshal(orig.Payload, &op); err != nil {
		t.Fatalf("decode orig payload: %v", err)
	}
	if dp.Content != op.Content {
		t.Errorf("payload mismatch: got=%+v want=%+v", dp, op)
	}
}

func TestJSONCodec_DecodeNotObject(t *testing.T) {
	codec := JSONCodec{}
	data := "[]" // JSON array, 首字符不是 '{'
	var buf bytes.Buffer
	buf.WriteString(data)

	decoded := &protocol.Envelope{}
	err := codec.Decode(&buf, decoded, 0)
	if err == nil {
		t.Errorf("Expected error for array input, got nil")
	}
}

func TestJSONCodec_DecodeMalformedJSON(t *testing.T) {
	codec := JSONCodec{}
	data := `{"type": "chat", "payload": {"content": "hello"}` // JSON 不完整
	var buf bytes.Buffer
	buf.WriteString(data)

	decoded := &protocol.Envelope{}
	err := codec.Decode(&buf, decoded, 0)
	if err == nil || !strings.Contains(err.Error(), "json decode") {
		t.Errorf("Expected json decode error, got %v", err)
	}
}

func TestJSONCodec_DecodeMissingType(t *testing.T) {
	codec := JSONCodec{}
	data := `{"payload": {"content": "hello"}}`
	var buf bytes.Buffer
	buf.WriteString(data)

	decoded := &protocol.Envelope{}
	err := codec.Decode(&buf, decoded, 0)
	if err == nil || !strings.Contains(err.Error(), "missing field: type") {
		t.Errorf("Expected missing field error, got %v", err)
	}
}

func TestJSONCodec_DecodeMaxSize(t *testing.T) {
	codec := JSONCodec{}
	data := `{"type": "chat", "content": "hello"}`
	var buf bytes.Buffer
	buf.WriteString(data)

	decoded := &protocol.Envelope{}
	// 设置 maxSize 小于实际长度
	err := codec.Decode(&buf, decoded, 5)
	if err == nil || !errors.Is(err, io.EOF) && !strings.Contains(err.Error(), "json decode") {
		t.Errorf("Expected decode error due to maxSize, got %v", err)
	}
}
