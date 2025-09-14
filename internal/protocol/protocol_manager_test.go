package protocol

import (
	"testing"
)

func TestProtocolManager_Creation(t *testing.T) {
	// Test creation with valid codec
	pm := NewProtocolManager(CodecJson)

	if pm.GetCodec() == nil {
		t.Error("Expected non-nil codec")
	}

	if pm.GetMessageFactory() == nil {
		t.Error("Expected non-nil message factory")
	}

	if pm.GetRouter() == nil {
		t.Error("Expected non-nil router")
	}
}

func TestProtocolManager_CodecFallback(t *testing.T) {
	// Test creation with invalid codec type (should fallback to JSON)
	pm := NewProtocolManager(999)

	codec := pm.GetCodec()
	if codec == nil {
		t.Error("Expected fallback codec")
	}

	// Should be JSON codec
	if codec.Name() != "json" {
		t.Errorf("Expected fallback to JSON codec, got %s", codec.Name())
	}
}

func TestProtocolManager_MessageHandlers(t *testing.T) {
	pm := NewProtocolManager(CodecJson)

	// Test registering a handler
	called := false
	handler := func(env *Envelope) error {
		called = true
		return nil
	}

	pm.RegisterMessageHandler(MsgText, handler)

	// Create a test envelope
	envelope := &Envelope{
		Type: MsgText,
		Data: []byte("test"),
	}

	// Dispatch should call our handler
	err := pm.GetRouter().Dispatch(envelope)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !called {
		t.Error("Expected handler to be called")
	}
}

func TestProtocolManager_DefaultHandler(t *testing.T) {
	pm := NewProtocolManager(CodecJson)

	// Set a default handler
	defaultCalled := false
	defaultHandler := func(env *Envelope) error {
		defaultCalled = true
		return nil
	}

	pm.SetDefaultMessageHandler(defaultHandler)

	// Create envelope with unregistered type
	envelope := &Envelope{
		Type: "unknown_type",
		Data: []byte("test"),
	}

	// Should call default handler
	err := pm.GetRouter().Dispatch(envelope)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !defaultCalled {
		t.Error("Expected default handler to be called")
	}
}
