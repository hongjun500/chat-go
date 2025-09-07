package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

// TestCodecFactory 测试编解码器工厂函数
func TestCodecFactory(t *testing.T) {
	tests := []struct {
		name      string
		codecType int8
		wantError bool
		wantType  string
	}{
		{"JSON Codec", 0, false, "json"},
		{"Protobuf Codec", 1, false, "protobuf"},
		{"Unknown Codec", 2, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codec, err := NewCodec(int(tt.codecType))

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for codec type %d, but got none", tt.codecType)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for codec type %d: %v", tt.codecType, err)
				return
			}

			if codec.Name() != tt.wantType {
				t.Errorf("Name mismatch: got %s, want %s", codec.Name(), tt.wantType)
			}
		})
	}
}

// TestCodecInteroperability 测试不同编码器的互操作性
func TestCodecInteroperability(t *testing.T) {
	// 创建测试消息
	envelope := &Envelope{
		Version:  "1.0",
		Type:     MsgText,
		Encoding: EncodingJSON,
		Mid:      "test-msg-001",
		From:     "alice",
		To:       "bob",
		Ts:       time.Now().UnixMilli(),
		Payload:  json.RawMessage(`{"text":"Hello World"}`),
	}

	// 测试 JSON 编解码
	t.Run("JSON Codec", func(t *testing.T) {
		jsonCodec, err := NewCodec(CodecJson)
		if err != nil {
			t.Fatalf("Failed to create JSON codec: %v", err)
		}

		// 编码
		var buf bytes.Buffer
		if err := jsonCodec.Encode(&buf, envelope); err != nil {
			t.Fatalf("JSON encode failed: %v", err)
		}

		// 解码
		var decoded Envelope
		if err := jsonCodec.Decode(&buf, &decoded, 1024*1024); err != nil {
			t.Fatalf("JSON decode failed: %v", err)
		}

		// 验证
		if decoded.Type != envelope.Type {
			t.Errorf("Type mismatch: got %s, want %s", decoded.Type, envelope.Type)
		}
		if decoded.Mid != envelope.Mid {
			t.Errorf("Mid mismatch: got %s, want %s", decoded.Mid, envelope.Mid)
		}
		if decoded.From != envelope.From {
			t.Errorf("From mismatch: got %s, want %s", decoded.From, envelope.From)
		}
	})

	// 测试 Protobuf 编解码
	t.Run("Protobuf Codec", func(t *testing.T) {
		protoCodec := &ProtobufCodec{}

		// 修改消息为 Protobuf 格式
		protoEnvelope := &Envelope{
			Version:  "1.0",
			Type:     MsgText,
			Encoding: EncodingProtobuf,
			Mid:      "test-msg-002",
			From:     "alice",
			To:       "bob",
			Ts:       time.Now().UnixMilli(),
			Data:     []byte("Hello World Protobuf"),
		}

		// 编码
		var buf bytes.Buffer
		if err := protoCodec.Encode(&buf, protoEnvelope); err != nil {
			t.Fatalf("Protobuf encode failed: %v", err)
		}

		// 解码
		var decoded Envelope
		if err := protoCodec.Decode(&buf, &decoded, 1024*1024); err != nil {
			t.Fatalf("Protobuf decode failed: %v", err)
		}

		// 验证
		if decoded.Type != protoEnvelope.Type {
			t.Errorf("Type mismatch: got %s, want %s", decoded.Type, protoEnvelope.Type)
		}
		if decoded.Mid != protoEnvelope.Mid {
			t.Errorf("Mid mismatch: got %s, want %s", decoded.Mid, protoEnvelope.Mid)
		}
		if decoded.From != protoEnvelope.From {
			t.Errorf("From mismatch: got %s, want %s", decoded.From, protoEnvelope.From)
		}
	})
}

// TestPayloadTypes 测试各种负载类型的序列化
func TestPayloadTypes(t *testing.T) {
	jsonCodec := &JSONCodec{}

	tests := []struct {
		name    string
		msgType MessageType
		payload interface{}
	}{
		{"Text Message", MsgText, TextPayload{Text: "Hello"}},
		{"Chat Message", "chat", ChatPayload{Content: "Hello chat"}},
		{"Set Name", "set_name", SetNamePayload{Name: "alice"}},
		{"Command", MsgCommand, CommandPayload{Raw: "/help"}},
		{"Ack", MsgAck, AckPayload{Status: "ok"}},
		{"Ping", MsgPing, PingPayload{Seq: 1, Timestamp: time.Now().UnixMilli()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 序列化负载
			payloadData, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			// 创建信封
			envelope := &Envelope{
				Type:    tt.msgType,
				Mid:     "test-" + string(tt.msgType),
				Ts:      time.Now().UnixMilli(),
				Payload: json.RawMessage(payloadData),
			}

			// 编码信封
			var buf bytes.Buffer
			if err := jsonCodec.Encode(&buf, envelope); err != nil {
				t.Fatalf("Failed to encode envelope: %v", err)
			}

			// 解码信封
			var decoded Envelope
			if err := jsonCodec.Decode(&buf, &decoded, 1024*1024); err != nil {
				t.Fatalf("Failed to decode envelope: %v", err)
			}

			// 验证
			if decoded.Type != envelope.Type {
				t.Errorf("Type mismatch: got %s, want %s", decoded.Type, envelope.Type)
			}
			if decoded.Mid != envelope.Mid {
				t.Errorf("Mid mismatch: got %s, want %s", decoded.Mid, envelope.Mid)
			}
		})
	}
}
