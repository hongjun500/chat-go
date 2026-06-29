package protocol

import (
	"testing"
	"time"
)

func TestMessageFactory_CreateTextMessage(t *testing.T) {
	factory := NewMessageFactory()
	
	msg := factory.CreateTextMessage("Hello World")
	
	if msg.Type != MsgText {
		t.Errorf("Expected message type %s, got %s", MsgText, msg.Type)
	}
	
	if msg.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", msg.Version)
	}
	
	if msg.Encoding != EncodingJSON {
		t.Errorf("Expected encoding %s, got %s", EncodingJSON, msg.Encoding)
	}
	
	if msg.Mid == "" {
		t.Error("Expected non-empty message ID")
	}
	
	if msg.Ts == 0 {
		t.Error("Expected non-zero timestamp")
	}
	
	if len(msg.Data) == 0 {
		t.Error("Expected non-empty data")
	}
}

func TestMessageFactory_CreateAckMessage(t *testing.T) {
	factory := NewMessageFactory()
	
	msg := factory.CreateAckMessage("ok", "test-correlation-id")
	
	if msg.Type != MsgAck {
		t.Errorf("Expected message type %s, got %s", MsgAck, msg.Type)
	}
	
	if msg.Correlation != "test-correlation-id" {
		t.Errorf("Expected correlation ID 'test-correlation-id', got %s", msg.Correlation)
	}
}

func TestMessageFactory_CreateCommandMessage(t *testing.T) {
	factory := NewMessageFactory()
	
	msg := factory.CreateCommandMessage("/help")
	
	if msg.Type != MsgCommand {
		t.Errorf("Expected message type %s, got %s", MsgCommand, msg.Type)
	}
}

func TestMessageFactory_CreatePingPongMessages(t *testing.T) {
	factory := NewMessageFactory()
	
	// Test ping
	ping := factory.CreatePingMessage(123)
	if ping.Type != MsgPing {
		t.Errorf("Expected ping message type %s, got %s", MsgPing, ping.Type)
	}
	
	// Test pong
	pong := factory.CreatePongMessage(123, ping.Mid)
	if pong.Type != MsgPong {
		t.Errorf("Expected pong message type %s, got %s", MsgPong, pong.Type)
	}
	
	if pong.Correlation != ping.Mid {
		t.Errorf("Expected pong correlation to match ping ID %s, got %s", ping.Mid, pong.Correlation)
	}
}

func TestMessageFactory_Timestamps(t *testing.T) {
	factory := NewMessageFactory()
	
	before := time.Now().UnixMilli()
	msg := factory.CreateTextMessage("test")
	after := time.Now().UnixMilli()
	
	if msg.Ts < before || msg.Ts > after {
		t.Errorf("Expected timestamp between %d and %d, got %d", before, after, msg.Ts)
	}
}