package transport

import (
	"strings"
	"testing"
	"time"

	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
	"github.com/hongjun500/chat-go/internal/protocol"
)

// mockTcpConn implements Session for testing (similar to tcpConn)
type mockTcpConn struct {
	id           string
	lastEnvelope *protocol.Envelope
	client       *chat.Client
}

func (m *mockTcpConn) ID() string                                          { return m.id }
func (m *mockTcpConn) RemoteAddr() string                                  { return "127.0.0.1:12345" }
func (m *mockTcpConn) SendEnvelope(env *protocol.Envelope) error           { m.lastEnvelope = env; return nil }
func (m *mockTcpConn) Close() error                                        { return nil }

func TestGatewayHandler_PayloadCodecIntegration(t *testing.T) {
	hub := chat.NewHub()
	cmdReg := command.NewRegistry()
	gateway := NewGatewayHandler(hub, cmdReg)

	// Test session (use mockTcpConn so getClient works)
	client := chat.NewClientWithBuffer("test", 10)
	session := &mockTcpConn{
		id:     "test-session",
		client: client,
	}

	// Test OnSessionOpen
	gateway.OnSessionOpen(session)
	if session.lastEnvelope == nil {
		t.Fatal("Expected welcome message envelope")
	}
	if session.lastEnvelope.Type != protocol.MsgText {
		t.Errorf("Expected text message type, got %s", session.lastEnvelope.Type)
	}

	// Decode the welcome message
	decoder := protocol.NewPayloadDecoder(protocol.DefaultPayloadCodec)
	payload, err := decoder.DecodeText(session.lastEnvelope)
	if err != nil {
		t.Fatalf("Failed to decode welcome message: %v", err)
	}
	if payload.Text != "请输入昵称并回车：" {
		t.Errorf("Unexpected welcome message: %s", payload.Text)
	}
}

func TestGatewayHandler_SetNameFlow(t *testing.T) {
	// Test set_name message encoding/decoding
	encoder := protocol.NewPayloadEncoder(protocol.DefaultPayloadCodec)
	setNameEnv, _ := encoder.EncodeSetName("TestUser")
	setNameEnv.Ts = time.Now().UnixMilli()

	// Verify that the payload codec works for encoding/decoding
	decoder := protocol.NewPayloadDecoder(protocol.DefaultPayloadCodec)
	payload, err := decoder.DecodeSetName(setNameEnv)
	if err != nil {
		t.Fatalf("Failed to decode set_name payload: %v", err)
	}
	if payload.Name != "TestUser" {
		t.Errorf("Expected name 'TestUser', got '%s'", payload.Name)
	}
}

func TestGatewayHandler_ChatMessage(t *testing.T) {
	// Test chat message encoding/decoding
	encoder := protocol.NewPayloadEncoder(protocol.DefaultPayloadCodec)
	chatEnv, _ := encoder.EncodeChat("Hello, World!")
	chatEnv.Ts = time.Now().UnixMilli()

	decoder := protocol.NewPayloadDecoder(protocol.DefaultPayloadCodec)
	payload, err := decoder.DecodeChat(chatEnv)
	if err != nil {
		t.Fatalf("Failed to decode chat payload: %v", err)
	}
	if payload.Content != "Hello, World!" {
		t.Errorf("Expected content 'Hello, World!', got '%s'", payload.Content)
	}
}

func TestGatewayHandler_PingPong(t *testing.T) {
	// Test ping/pong message encoding/decoding
	codec := protocol.NewPayloadCodec()
	pingEnv, _ := codec.CreateEnvelope(protocol.MsgPing, protocol.PingPayload{Seq: 123})
	pingEnv.Ts = time.Now().UnixMilli()

	decoder := protocol.NewPayloadDecoder(protocol.DefaultPayloadCodec)
	payload, err := decoder.DecodePing(pingEnv)
	if err != nil {
		t.Fatalf("Failed to decode ping payload: %v", err)
	}
	if payload.Seq != 123 {
		t.Errorf("Expected seq 123, got %d", payload.Seq)
	}
}

func TestGatewayHandler_UnknownMessageType(t *testing.T) {
	// Test that unknown message types are handled properly
	codec := protocol.NewPayloadCodec()
	unknownEnv := &protocol.Envelope{
		Type:    "unknown_type",
		Payload: []byte(`{"test": "data"}`),
		Ts:      time.Now().UnixMilli(),
	}
	
	_, err := codec.DecodePayloadByType(unknownEnv)
	if err == nil {
		t.Error("Expected error for unknown message type")
	}
	if !strings.Contains(err.Error(), "unknown message type") {
		t.Errorf("Expected 'unknown message type' error, got: %v", err)
	}
}