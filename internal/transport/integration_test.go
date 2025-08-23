package transport

import (
	"testing"

	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
)

func TestTransportIntegration_Codecs(t *testing.T) {
	// Test that both TCP and WebSocket can use different codecs
	hub := chat.NewHub()
	cmdReg := command.NewRegistry()
	_ = &GatewayHandler{Hub: hub, Commands: cmdReg} // Create gateway for potential future use

	// Test codec compatibility
	t.Run("TCP_Uses_JSON_Codec", func(t *testing.T) {
		tcpSrv := &TCPServer{Codec: &JSONCodec{}}
		if tcpSrv.Name() != Tcp {
			t.Errorf("Expected TCP name, got %s", tcpSrv.Name())
		}
		if tcpSrv.Codec.ContentType() != ApplicationJson {
			t.Errorf("Expected JSON content type, got %s", tcpSrv.Codec.ContentType())
		}
	})

	t.Run("TCP_Uses_Protobuf_Codec", func(t *testing.T) {
		tcpSrv := &TCPServer{Codec: &ProtobufCodec{}}
		if tcpSrv.Codec.ContentType() != ApplicationProtobuf {
			t.Errorf("Expected Protobuf content type, got %s", tcpSrv.Codec.ContentType())
		}
	})

	t.Run("WebSocket_Uses_JSON_Codec", func(t *testing.T) {
		wsSrv := &WebSocketServer{Codec: &JSONCodec{}}
		if wsSrv.Name() != WebSocket {
			t.Errorf("Expected WebSocket name, got %s", wsSrv.Name())
		}
		if wsSrv.Codec.ContentType() != ApplicationJson {
			t.Errorf("Expected JSON content type, got %s", wsSrv.Codec.ContentType())
		}
	})

	t.Run("WebSocket_Uses_Protobuf_Codec", func(t *testing.T) {
		wsSrv := &WebSocketServer{Codec: &ProtobufCodec{}}
		if wsSrv.Codec.ContentType() != ApplicationProtobuf {
			t.Errorf("Expected Protobuf content type, got %s", wsSrv.Codec.ContentType())
		}
	})

	// Test Transport interface compliance
	t.Run("Both_Implement_Transport", func(t *testing.T) {
		var tcpTransport Transport = &TCPServer{}
		var wsTransport Transport = &WebSocketServer{}
		
		if tcpTransport.Name() != Tcp {
			t.Errorf("TCP transport name mismatch")
		}
		if wsTransport.Name() != WebSocket {
			t.Errorf("WebSocket transport name mismatch")
		}
	})
}

