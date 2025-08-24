package protocol

import (
	"encoding/json"
	"testing"
)

func TestPayloadCodec_EncodeDecodeText(t *testing.T) {
	codec := NewPayloadCodec()
	decoder := NewPayloadDecoder(codec)
	encoder := NewPayloadEncoder(codec)

	// Test encoding
	envelope, err := encoder.EncodeText("Hello, World!")
	if err != nil {
		t.Fatalf("Failed to encode text: %v", err)
	}

	if envelope.Type != MsgText {
		t.Errorf("Expected type %s, got %s", MsgText, envelope.Type)
	}

	// Test decoding
	payload, err := decoder.DecodeText(envelope)
	if err != nil {
		t.Fatalf("Failed to decode text: %v", err)
	}

	if payload.Text != "Hello, World!" {
		t.Errorf("Expected text 'Hello, World!', got '%s'", payload.Text)
	}
}

func TestPayloadCodec_EncodeDecodeChat(t *testing.T) {
	codec := NewPayloadCodec()
	decoder := NewPayloadDecoder(codec)
	encoder := NewPayloadEncoder(codec)

	// Test encoding
	envelope, err := encoder.EncodeChat("Test chat message")
	if err != nil {
		t.Fatalf("Failed to encode chat: %v", err)
	}

	if envelope.Type != "chat" {
		t.Errorf("Expected type 'chat', got %s", envelope.Type)
	}

	// Test decoding
	payload, err := decoder.DecodeChat(envelope)
	if err != nil {
		t.Fatalf("Failed to decode chat: %v", err)
	}

	if payload.Content != "Test chat message" {
		t.Errorf("Expected content 'Test chat message', got '%s'", payload.Content)
	}
}

func TestPayloadCodec_EncodeDecodeCommand(t *testing.T) {
	codec := NewPayloadCodec()
	decoder := NewPayloadDecoder(codec)
	encoder := NewPayloadEncoder(codec)

	// Test encoding
	envelope, err := encoder.EncodeCommand("/help")
	if err != nil {
		t.Fatalf("Failed to encode command: %v", err)
	}

	if envelope.Type != MsgCommand {
		t.Errorf("Expected type %s, got %s", MsgCommand, envelope.Type)
	}

	// Test decoding
	payload, err := decoder.DecodeCommand(envelope)
	if err != nil {
		t.Fatalf("Failed to decode command: %v", err)
	}

	if payload.Raw != "/help" {
		t.Errorf("Expected raw '/help', got '%s'", payload.Raw)
	}
}

func TestPayloadCodec_DecodePayloadByType(t *testing.T) {
	codec := NewPayloadCodec()

	// Create a text envelope manually
	textPayload := TextPayload{Text: "Test message"}
	payloadBytes, _ := json.Marshal(textPayload)
	envelope := &Envelope{
		Type:    MsgText,
		Payload: json.RawMessage(payloadBytes),
	}

	// Decode using type-based decoding
	decodedPayload, err := codec.DecodePayloadByType(envelope)
	if err != nil {
		t.Fatalf("Failed to decode payload by type: %v", err)
	}

	// Type assertion
	textResult, ok := decodedPayload.(TextPayload)
	if !ok {
		t.Fatalf("Expected TextPayload, got %T", decodedPayload)
	}

	if textResult.Text != "Test message" {
		t.Errorf("Expected text 'Test message', got '%s'", textResult.Text)
	}
}

func TestPayloadCodec_UnknownType(t *testing.T) {
	codec := NewPayloadCodec()

	envelope := &Envelope{
		Type:    "unknown_type",
		Payload: json.RawMessage(`{"test": "data"}`),
	}

	_, err := codec.DecodePayloadByType(envelope)
	if err == nil {
		t.Error("Expected error for unknown message type")
	}
}

func TestPayloadCodec_EmptyPayload(t *testing.T) {
	codec := NewPayloadCodec()

	envelope := &Envelope{
		Type:    MsgText,
		Payload: nil,
	}

	var payload TextPayload
	err := codec.DecodePayload(envelope, &payload)
	if err == nil {
		t.Error("Expected error for empty payload")
	}
}

func TestPayloadCodec_InvalidJSON(t *testing.T) {
	codec := NewPayloadCodec()

	envelope := &Envelope{
		Type:    MsgText,
		Payload: json.RawMessage(`{invalid json`),
	}

	var payload TextPayload
	err := codec.DecodePayload(envelope, &payload)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestPayloadEncoder_EncodeDirect(t *testing.T) {
	encoder := NewPayloadEncoder(NewPayloadCodec())

	envelope, err := encoder.EncodeDirect([]string{"user1", "user2"}, "Private message")
	if err != nil {
		t.Fatalf("Failed to encode direct message: %v", err)
	}

	if envelope.Type != "direct" {
		t.Errorf("Expected type 'direct', got %s", envelope.Type)
	}

	// Decode to verify
	var payload DirectPayload
	err = json.Unmarshal(envelope.Payload, &payload)
	if err != nil {
		t.Fatalf("Failed to decode direct payload: %v", err)
	}

	if len(payload.To) != 2 || payload.To[0] != "user1" || payload.To[1] != "user2" {
		t.Errorf("Expected To: [user1, user2], got %v", payload.To)
	}

	if payload.Content != "Private message" {
		t.Errorf("Expected content 'Private message', got '%s'", payload.Content)
	}
}

func TestPayloadEncoder_EncodeAck(t *testing.T) {
	encoder := NewPayloadEncoder(NewPayloadCodec())

	envelope, err := encoder.EncodeAck("ok")
	if err != nil {
		t.Fatalf("Failed to encode ack: %v", err)
	}

	if envelope.Type != MsgAck {
		t.Errorf("Expected type %s, got %s", MsgAck, envelope.Type)
	}

	// Decode to verify
	var payload AckPayload
	err = json.Unmarshal(envelope.Payload, &payload)
	if err != nil {
		t.Fatalf("Failed to decode ack payload: %v", err)
	}

	if payload.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", payload.Status)
	}
}

func TestMustJSON_BackwardCompatibility(t *testing.T) {
	// Test that MustJSON works like the old mustJSON function
	payload := TextPayload{Text: "test"}
	
	// This should not panic
	result := MustJSON(payload)
	
	// Verify it produces valid JSON
	var decoded TextPayload
	err := json.Unmarshal(result, &decoded)
	if err != nil {
		t.Fatalf("MustJSON produced invalid JSON: %v", err)
	}
	
	if decoded.Text != "test" {
		t.Errorf("Expected text 'test', got '%s'", decoded.Text)
	}
}

func TestPayloadCodec_SafeDecodePayload(t *testing.T) {
	codec := NewPayloadCodec()
	
	// Valid payload
	envelope, _ := DefaultPayloadEncoder.EncodeText("test")
	var payload TextPayload
	if !codec.SafeDecodePayload(envelope, &payload) {
		t.Error("SafeDecodePayload should return true for valid payload")
	}
	
	// Invalid payload
	envelope.Payload = json.RawMessage(`{invalid json`)
	if codec.SafeDecodePayload(envelope, &payload) {
		t.Error("SafeDecodePayload should return false for invalid payload")
	}
}