package protocol

import (
	"testing"
)

// TestMessageFactory 测试消息工厂
func TestMessageFactory(t *testing.T) {
	jsonCodec := &JSONCodec{}
	config := NewDefaultCodecConfig(CodecJson)
	factory := NewMessageFactory(jsonCodec, config)
	
	// 测试创建文本消息
	textMsg, err := factory.CreateTextMessage("Hello World", "alice", "bob")
	if err != nil {
		t.Fatalf("Failed to create text message: %v", err)
	}
	
	if textMsg.Type != MsgText {
		t.Errorf("Expected message type %s, got %s", MsgText, textMsg.Type)
	}
	if textMsg.From != "alice" {
		t.Errorf("Expected from 'alice', got '%s'", textMsg.From)
	}
	if textMsg.To != "bob" {
		t.Errorf("Expected to 'bob', got '%s'", textMsg.To)
	}
	if len(textMsg.Data) == 0 {
		t.Error("Message data should not be empty")
	}
	
	// 测试创建聊天消息
	chatMsg, err := factory.CreateChatMessage("Chat content", "alice")
	if err != nil {
		t.Fatalf("Failed to create chat message: %v", err)
	}
	
	if chatMsg.Type != "chat" {
		t.Errorf("Expected message type 'chat', got %s", chatMsg.Type)
	}
	if chatMsg.From != "alice" {
		t.Errorf("Expected from 'alice', got '%s'", chatMsg.From)
	}
	
	// 测试创建确认消息
	ackMsg, err := factory.CreateAckMessage("ok", "test-correlation")
	if err != nil {
		t.Fatalf("Failed to create ack message: %v", err)
	}
	
	if ackMsg.Type != MsgAck {
		t.Errorf("Expected message type %s, got %s", MsgAck, ackMsg.Type)
	}
	if ackMsg.Correlation != "test-correlation" {
		t.Errorf("Expected correlation 'test-correlation', got '%s'", ackMsg.Correlation)
	}
	
	// 测试创建命令消息
	cmdMsg, err := factory.CreateCommandMessage("/help", "alice")
	if err != nil {
		t.Fatalf("Failed to create command message: %v", err)
	}
	
	if cmdMsg.Type != MsgCommand {
		t.Errorf("Expected message type %s, got %s", MsgCommand, cmdMsg.Type)
	}
	if cmdMsg.From != "alice" {
		t.Errorf("Expected from 'alice', got '%s'", cmdMsg.From)
	}
}

// TestMessageFactoryEncoding 测试消息工厂编码功能
func TestMessageFactoryEncoding(t *testing.T) {
	jsonCodec := &JSONCodec{}
	config := NewDefaultCodecConfig(CodecJson)
	factory := NewMessageFactory(jsonCodec, config)
	
	// 创建测试消息
	msg, err := factory.CreateTextMessage("Test message", "alice", "bob")
	if err != nil {
		t.Fatalf("Failed to create test message: %v", err)
	}
	
	// 测试编码
	data, err := factory.EncodeMessage(msg)
	if err != nil {
		t.Fatalf("Failed to encode message: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Encoded data should not be empty")
	}
	
	// 测试解码
	decoded, err := factory.DecodeMessage(data, 1024*1024)
	if err != nil {
		t.Fatalf("Failed to decode message: %v", err)
	}
	
	if decoded.Type != msg.Type {
		t.Errorf("Type mismatch: got %s, want %s", decoded.Type, msg.Type)
	}
	if decoded.From != msg.From {
		t.Errorf("From mismatch: got %s, want %s", decoded.From, msg.From)
	}
	if decoded.To != msg.To {
		t.Errorf("To mismatch: got %s, want %s", decoded.To, msg.To)
	}
	
	// 测试使用指定编码器解码
	decoded2, err := factory.DecodeMessageWithCodec(data, CodecJson, 1024*1024)
	if err != nil {
		t.Fatalf("Failed to decode with specific codec: %v", err)
	}
	
	if decoded2.Type != msg.Type {
		t.Errorf("Type mismatch with specific codec: got %s, want %s", decoded2.Type, msg.Type)
	}
}

// TestMessageFactoryErrors 测试消息工厂错误处理
func TestMessageFactoryErrors(t *testing.T) {
	jsonCodec := &JSONCodec{}
	config := NewDefaultCodecConfig(CodecJson)
	factory := NewMessageFactory(jsonCodec, config)
	
	// 测试创建空文本消息
	_, err := factory.CreateTextMessage("", "alice", "bob")
	if err == nil {
		t.Error("Expected error when creating text message with empty text")
	}
	
	// 测试创建空聊天消息
	_, err = factory.CreateChatMessage("", "alice")
	if err == nil {
		t.Error("Expected error when creating chat message with empty content")
	}
	
	// 测试创建空确认消息
	_, err = factory.CreateAckMessage("", "correlation")
	if err == nil {
		t.Error("Expected error when creating ack message with empty status")
	}
	
	// 测试创建空命令消息
	_, err = factory.CreateCommandMessage("", "alice")
	if err == nil {
		t.Error("Expected error when creating command message with empty command")
	}
	
	// 测试编码空消息
	_, err = factory.EncodeMessage(nil)
	if err == nil {
		t.Error("Expected error when encoding nil message")
	}
	
	// 测试解码空数据
	_, err = factory.DecodeMessage([]byte{}, 1024)
	if err == nil {
		t.Error("Expected error when decoding empty data")
	}
	
	// 测试解码超过最大大小的数据
	largeData := make([]byte, 1000)
	_, err = factory.DecodeMessage(largeData, 100)
	if err == nil {
		t.Error("Expected error when decoding data exceeding max size")
	}
	
	// 测试使用不支持的编码器解码
	testData := []byte(`{"type":"text","mid":"test"}`)
	_, err = factory.DecodeMessageWithCodec(testData, 999, 1024)
	if err == nil {
		t.Error("Expected error when using unsupported codec")
	}
}

// TestMessageFactoryWithVersionControl 测试带版本控制的消息工厂
func TestMessageFactoryWithVersionControl(t *testing.T) {
	jsonCodec := &JSONCodec{}
	config := NewDefaultCodecConfig(CodecJson)
	vc := NewVersionController("1.0")
	
	// 注册版本映射
	codecMapping := CodecMapping{
		DefaultCodec: CodecJson,
		Codecs:       make(map[MessageType]int),
	}
	handlerMapping := HandlerMapping{
		DefaultHandler: func(env *Envelope) error { return nil },
		Handlers:       make(map[MessageType]MessageHandler),
	}
	
	err := vc.RegisterVersion("1.0", codecMapping, handlerMapping)
	if err != nil {
		t.Fatalf("Failed to register version: %v", err)
	}
	
	factory := NewMessageFactoryWithVersionControl(jsonCodec, config, vc)
	
	// 创建消息
	msg, err := factory.CreateTextMessage("Test with version control", "alice", "bob")
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}
	
	// 编码消息
	data, err := factory.EncodeMessage(msg)
	if err != nil {
		t.Fatalf("Failed to encode message: %v", err)
	}
	
	// 解码消息
	decoded, err := factory.DecodeMessage(data, 1024*1024)
	if err != nil {
		t.Fatalf("Failed to decode message: %v", err)
	}
	
	if decoded.Type != msg.Type {
		t.Errorf("Type mismatch: got %s, want %s", decoded.Type, msg.Type)
	}
	
	// 测试设置和获取默认编码器
	protoCodec := &ProtobufCodec{}
	factory.SetDefaultCodec(protoCodec)
	
	if factory.GetDefaultCodec().Name() != protoCodec.Name() {
		t.Errorf("Expected default codec %s, got %s", protoCodec.Name(), factory.GetDefaultCodec().Name())
	}
}