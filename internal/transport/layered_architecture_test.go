package transport

import (
	"testing"
	"github.com/hongjun500/chat-go/internal/protocol"
)

func TestSimpleGateway_Creation(t *testing.T) {
	pm := protocol.NewProtocolManager(protocol.CodecJson)
	gw := NewSimpleGateway(pm)
	
	if gw.GetProtocolManager() != pm {
		t.Error("Expected protocol manager to be set correctly")
	}
	
	if gw.GetSessionManager() == nil {
		t.Error("Expected non-nil session manager")
	}
}

func TestSimpleGateway_BackwardCompatibility(t *testing.T) {
	// Test backward compatible constructor
	gw := NewSimpleGatewayWithCodec(protocol.CodecJson)
	
	if gw.GetProtocolManager() == nil {
		t.Error("Expected non-nil protocol manager from backward compatible constructor")
	}
	
	if gw.GetProtocolManager().GetCodec().Name() != "json" {
		t.Errorf("Expected JSON codec, got %s", gw.GetProtocolManager().GetCodec().Name())
	}
}

func TestSimpleGateway_MessageHandlers(t *testing.T) {
	pm := protocol.NewProtocolManager(protocol.CodecJson)
	gw := NewSimpleGateway(pm)
	
	// Test registering session-level handler
	called := false
	gw.RegisterHandler("test", func(sess Session, env *protocol.Envelope) {
		called = true
	})
	
	// Test registering protocol-level handler
	protocolCalled := false
	gw.RegisterProtocolHandler(protocol.MsgText, func(env *protocol.Envelope) error {
		protocolCalled = true
		return nil
	})
	
	// Create test envelope for session handler
	env1 := &protocol.Envelope{Type: "test"}
	gw.OnEnvelope(nil, env1) // Session can be nil for this test
	
	if !called {
		t.Error("Expected session handler to be called")
	}
	
	// Create test envelope for protocol handler
	env2 := &protocol.Envelope{Type: protocol.MsgText}
	gw.OnEnvelope(nil, env2)
	
	if !protocolCalled {
		t.Error("Expected protocol handler to be called")
	}
}

func TestTransportOptions_ProtocolManagerGetters(t *testing.T) {
	// Create protocol managers
	tcpPM := protocol.NewProtocolManager(protocol.CodecJson)
	wsPM := protocol.NewProtocolManager(protocol.CodecProtobuf)
	
	// Test with explicit protocol managers
	opts := Options{
		TCPProtocolManager: tcpPM,
		WSProtocolManager:  wsPM,
	}
	
	if opts.GetTCPProtocolManager() != tcpPM {
		t.Error("Expected TCP protocol manager to be returned")
	}
	
	if opts.GetWSProtocolManager() != wsPM {
		t.Error("Expected WS protocol manager to be returned")
	}
	
	// Test backward compatibility (fallback to codec types)
	optsCompat := Options{
		TCPCodec: protocol.CodecJson,
		WSCodec:  protocol.CodecProtobuf,
	}
	
	tcpPMCompat := optsCompat.GetTCPProtocolManager()
	if tcpPMCompat == nil {
		t.Error("Expected TCP protocol manager from backward compatibility")
	}
	
	if tcpPMCompat.GetCodec().Name() != "json" {
		t.Errorf("Expected JSON codec from backward compatibility, got %s", tcpPMCompat.GetCodec().Name())
	}
	
	wsPMCompat := optsCompat.GetWSProtocolManager()
	if wsPMCompat == nil {
		t.Error("Expected WS protocol manager from backward compatibility")
	}
	
	if wsPMCompat.GetCodec().Name() != "protobuf" {
		t.Errorf("Expected protobuf codec from backward compatibility, got %s", wsPMCompat.GetCodec().Name())
	}
}

func TestSessionManager_Operations(t *testing.T) {
	sm := NewSessionManager()
	
	// Create a mock session
	mockSession := &mockSession{id: "test-session", addr: "127.0.0.1:8080"}
	
	// Test adding session
	sm.AddSession(mockSession)
	
	// Test getting session
	session, exists := sm.GetSession("test-session")
	if !exists {
		t.Error("Expected session to exist")
	}
	
	if session.ID() != "test-session" {
		t.Errorf("Expected session ID 'test-session', got %s", session.ID())
	}
	
	// Test getting all sessions
	allSessions := sm.GetAllSessions()
	if len(allSessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(allSessions))
	}
	
	// Test removing session
	sm.RemoveSession("test-session")
	
	_, exists = sm.GetSession("test-session")
	if exists {
		t.Error("Expected session to be removed")
	}
}

// Mock session for testing
type mockSession struct {
	id   string
	addr string
}

func (m *mockSession) ID() string {
	return m.id
}

func (m *mockSession) RemoteAddr() string {
	return m.addr
}

func (m *mockSession) SendEnvelope(env *protocol.Envelope) error {
	return nil
}

func (m *mockSession) Close() error {
	return nil
}