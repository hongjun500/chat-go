package protocol

import (
	"testing"
)

// TestVersionController 测试版本控制器
func TestVersionController(t *testing.T) {
	vc := NewVersionController("1.0")
	
	// 测试默认版本
	if vc.GetDefaultVersion() != "1.0" {
		t.Errorf("Expected default version '1.0', got '%s'", vc.GetDefaultVersion())
	}
	
	// 注册版本映射
	codecMapping := CodecMapping{
		DefaultCodec: CodecJson,
		Codecs: map[MessageType]int{
			MsgText: CodecProtobuf,
		},
	}
	
	handlerMapping := HandlerMapping{
		DefaultHandler: func(env *Envelope) error { return nil },
		Handlers: map[MessageType]MessageHandler{
			MsgText: func(env *Envelope) error { return nil },
		},
	}
	
	err := vc.RegisterVersion("1.1", codecMapping, handlerMapping)
	if err != nil {
		t.Fatalf("Failed to register version: %v", err)
	}
	
	// 测试版本支持检查
	if !vc.IsVersionSupported("1.1") {
		t.Error("Version 1.1 should be supported")
	}
	
	if vc.IsVersionSupported("2.0") {
		t.Error("Version 2.0 should not be supported")
	}
	
	// 测试获取支持的版本
	versions := vc.GetSupportedVersions()
	found := false
	for _, v := range versions {
		if v == "1.1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Version 1.1 should be in supported versions list")
	}
	
	// 测试获取编码器
	envelope := &Envelope{
		Version: "1.1",
		Type:    MsgText,
	}
	
	codec, err := vc.GetCodecForMessage(envelope)
	if err != nil {
		t.Fatalf("Failed to get codec for message: %v", err)
	}
	
	if codec.Name() != Protobuf {
		t.Errorf("Expected Protobuf codec for MsgText in version 1.1, got %s", codec.Name())
	}
	
	// 测试获取处理器
	handler, err := vc.GetHandlerForMessage(envelope)
	if err != nil {
		t.Fatalf("Failed to get handler for message: %v", err)
	}
	
	if handler == nil {
		t.Error("Handler should not be nil")
	}
	
	// 测试默认编码器
	envelope.Type = MsgCommand // 没有特定映射的消息类型
	codec, err = vc.GetCodecForMessage(envelope)
	if err != nil {
		t.Fatalf("Failed to get default codec: %v", err)
	}
	
	if codec.Name() != Json {
		t.Errorf("Expected JSON default codec, got %s", codec.Name())
	}
	
	// 测试设置默认版本
	err = vc.SetDefaultVersion("1.1")
	if err != nil {
		t.Fatalf("Failed to set default version: %v", err)
	}
	
	if vc.GetDefaultVersion() != "1.1" {
		t.Errorf("Expected default version '1.1', got '%s'", vc.GetDefaultVersion())
	}
	
	// 测试设置不存在的默认版本
	err = vc.SetDefaultVersion("2.0")
	if err == nil {
		t.Error("Expected error when setting non-existent default version")
	}
}

// TestVersionedProtocol 测试版本化协议
func TestVersionedProtocol(t *testing.T) {
	vp := NewVersionedProtocol(CodecJson, "1.0")
	
	// 测试注册版本映射
	codecMapping := CodecMapping{
		DefaultCodec: CodecProtobuf,
		Codecs:       make(map[MessageType]int),
	}
	
	handlerMapping := HandlerMapping{
		DefaultHandler: func(env *Envelope) error { return nil },
		Handlers:       make(map[MessageType]MessageHandler),
	}
	
	err := vp.RegisterVersionMapping("1.1", codecMapping, handlerMapping)
	if err != nil {
		t.Fatalf("Failed to register version mapping: %v", err)
	}
	
	// 测试版本化分发
	envelope := &Envelope{
		Version: "1.1",
		Type:    MsgText,
	}
	
	err = vp.DispatchVersioned(envelope)
	if err != nil {
		t.Fatalf("Failed to dispatch versioned message: %v", err)
	}
	
	// 测试获取版本控制器
	vc := vp.GetVersionController()
	if vc == nil {
		t.Error("Version controller should not be nil")
	}
	
	if !vc.IsVersionSupported("1.1") {
		t.Error("Version 1.1 should be supported by version controller")
	}
}

// TestVersionControllerErrors 测试版本控制器错误情况
func TestVersionControllerErrors(t *testing.T) {
	vc := NewVersionController("1.0")
	
	// 测试注册不支持的编码器
	codecMapping := CodecMapping{
		DefaultCodec: 999, // 不支持的编码器
		Codecs:       make(map[MessageType]int),
	}
	
	handlerMapping := HandlerMapping{
		DefaultHandler: func(env *Envelope) error { return nil },
		Handlers:       make(map[MessageType]MessageHandler),
	}
	
	err := vc.RegisterVersion("1.1", codecMapping, handlerMapping)
	if err == nil {
		t.Error("Expected error when registering version with unsupported codec")
	}
	
	// 测试注册消息类型特定的不支持编码器
	codecMapping.DefaultCodec = CodecJson
	codecMapping.Codecs[MsgText] = 999
	
	err = vc.RegisterVersion("1.2", codecMapping, handlerMapping)
	if err == nil {
		t.Error("Expected error when registering version with unsupported message-specific codec")
	}
	
	// 测试获取不存在版本的编码器
	envelope := &Envelope{
		Version: "2.0",
		Type:    MsgText,
	}
	
	_, err = vc.GetCodecForMessage(envelope)
	if err == nil {
		t.Error("Expected error when getting codec for unsupported version")
	}
	
	// 测试获取不存在版本的处理器
	_, err = vc.GetHandlerForMessage(envelope)
	if err == nil {
		t.Error("Expected error when getting handler for unsupported version")
	}
}