package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
)

func TestWebSocketServer_Transport(t *testing.T) {
	// Create test server
	hub := chat.NewHub()
	cmdReg := command.NewRegistry()
	wsSrv := &WebSocketServer{Codec: &JSONCodec{}}
	gw := &GatewayHandler{Hub: hub, Commands: cmdReg}

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsSrv.handleConnection(w, r, gw, Options{OutBuffer: 10})
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	u, _ := url.Parse(wsURL)

	// Test WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read initial greeting
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read greeting: %v", err)
	}

	var greeting Envelope
	if err := json.Unmarshal(message, &greeting); err != nil {
		t.Fatalf("Failed to unmarshal greeting: %v", err)
	}

	if greeting.Type != "text" {
		t.Errorf("Expected greeting type 'text', got %s", greeting.Type)
	}

	var textPayload TextPayload
	if err := json.Unmarshal(greeting.Payload, &textPayload); err != nil {
		t.Fatalf("Failed to unmarshal text payload: %v", err)
	}

	if !strings.Contains(textPayload.Text, "昵称") {
		t.Errorf("Expected greeting to contain '昵称', got: %s", textPayload.Text)
	}
}

func TestWebSocketServer_LegacyTextMessage(t *testing.T) {
	// Create test server
	hub := chat.NewHub()
	cmdReg := command.NewRegistry()
	wsSrv := &WebSocketServer{Codec: &JSONCodec{}}
	gw := &GatewayHandler{Hub: hub, Commands: cmdReg}

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsSrv.handleConnection(w, r, gw, Options{OutBuffer: 10})
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	u, _ := url.Parse(wsURL)

	// Test WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read and discard greeting
	_, _, _ = conn.ReadMessage()

	// Send plain text message (legacy format) to set name
	if err := conn.WriteMessage(websocket.TextMessage, []byte("testuser")); err != nil {
		t.Fatalf("Failed to send name: %v", err)
	}

	// Read ack message
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read ack: %v", err)
	}

	var ack Envelope
	if err := json.Unmarshal(message, &ack); err != nil {
		t.Fatalf("Failed to unmarshal ack: %v", err)
	}

	if ack.Type != "ack" {
		t.Errorf("Expected ack type, got %s", ack.Type)
	}
}

func TestWebSocketServer_StructuredMessage(t *testing.T) {
	// Create test server
	hub := chat.NewHub()
	cmdReg := command.NewRegistry()
	wsSrv := &WebSocketServer{Codec: &JSONCodec{}}
	gw := &GatewayHandler{Hub: hub, Commands: cmdReg}

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsSrv.handleConnection(w, r, gw, Options{OutBuffer: 10})
	}))
	defer server.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	u, _ := url.Parse(wsURL)

	// Test WebSocket connection
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read and discard greeting
	_, _, _ = conn.ReadMessage()

	// Send structured set_name message
	payload, _ := json.Marshal(SetNamePayload{Name: "testuser"})
	envelope := &Envelope{
		Type:    "set_name",
		Payload: payload,
		Ts:      time.Now().UnixMilli(),
	}

	data, _ := json.Marshal(envelope)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("Failed to send structured message: %v", err)
	}

	// Read ack message
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read ack: %v", err)
	}

	var ack Envelope
	if err := json.Unmarshal(message, &ack); err != nil {
		t.Fatalf("Failed to unmarshal ack: %v", err)
	}

	if ack.Type != "ack" {
		t.Errorf("Expected ack type, got %s", ack.Type)
	}
}

func TestWebSocketServer_ImplementsTransport(t *testing.T) {
	var _ Transport = &WebSocketServer{}
	
	wsSrv := &WebSocketServer{}
	if wsSrv.Name() != WebSocket {
		t.Errorf("Expected WebSocket name, got %s", wsSrv.Name())
	}
}