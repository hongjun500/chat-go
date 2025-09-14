package protocol

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"
)

// TestGenericCodec 测试通用编码器
func TestGenericCodec(t *testing.T) {
	// 创建一个简单的编码器
	encoder := func(w io.Writer, e *Envelope) error {
		return json.NewEncoder(w).Encode(e)
	}
	
	decoder := func(r io.Reader, e *Envelope, maxSize int) error {
		return json.NewDecoder(r).Decode(e)
	}
	
	codec := NewGenericCodec("test", encoder, decoder)
	
	// 测试编码器名称
	if codec.Name() != "test" {
		t.Errorf("Expected codec name 'test', got '%s'", codec.Name())
	}
	
	// 测试编码
	envelope := &Envelope{
		Version: "1.0",
		Type:    MsgText,
		Mid:     "test-msg",
		Ts:      time.Now().UnixMilli(),
		Data:    []byte("test data"),
	}
	
	var buf bytes.Buffer
	err := codec.Encode(&buf, envelope)
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}
	
	// 测试解码
	var decoded Envelope
	err = codec.Decode(&buf, &decoded, 1024)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	
	// 验证结果
	if decoded.Type != envelope.Type {
		t.Errorf("Type mismatch: got %s, want %s", decoded.Type, envelope.Type)
	}
	if decoded.Mid != envelope.Mid {
		t.Errorf("Mid mismatch: got %s, want %s", decoded.Mid, envelope.Mid)
	}
	
	// 测试错误情况
	err = codec.Encode(nil, envelope)
	if err == nil {
		t.Error("Expected error when encoding with nil writer")
	}
	
	err = codec.Encode(&buf, nil)
	if err == nil {
		t.Error("Expected error when encoding nil envelope")
	}
	
	err = codec.Decode(nil, &decoded, 1024)
	if err == nil {
		t.Error("Expected error when decoding with nil reader")
	}
	
	err = codec.Decode(&buf, nil, 1024)
	if err == nil {
		t.Error("Expected error when decoding to nil envelope")
	}
}

// TestCodecWrapper 测试编码器包装器
func TestCodecWrapper(t *testing.T) {
	jsonCodec := &JSONCodec{}
	wrapper := NewCodecWrapper(jsonCodec)
	
	// 测试包装器名称
	if wrapper.Name() != jsonCodec.Name() {
		t.Errorf("Expected wrapper name '%s', got '%s'", jsonCodec.Name(), wrapper.Name())
	}
	
	// 测试编码
	envelope := &Envelope{
		Version: "1.0",
		Type:    MsgText,
		Mid:     "test-msg",
		Ts:      time.Now().UnixMilli(),
		Data:    []byte("test data"),
	}
	
	var buf bytes.Buffer
	err := wrapper.Encode(&buf, envelope)
	if err != nil {
		t.Fatalf("Failed to encode with wrapper: %v", err)
	}
	
	// 测试解码
	var decoded Envelope
	err = wrapper.Decode(&buf, &decoded, 1024)
	if err != nil {
		t.Fatalf("Failed to decode with wrapper: %v", err)
	}
	
	// 验证结果
	if decoded.Type != envelope.Type {
		t.Errorf("Type mismatch: got %s, want %s", decoded.Type, envelope.Type)
	}
	
	// 测试错误情况
	err = wrapper.Encode(nil, envelope)
	if err == nil {
		t.Error("Expected error when encoding with nil writer")
	}
	
	err = wrapper.Decode(nil, &decoded, 1024)
	if err == nil {
		t.Error("Expected error when decoding with nil reader")
	}
}

// TestEnhancedErrorHandling 测试增强的错误处理
func TestEnhancedErrorHandling(t *testing.T) {
	// 测试JSON编码器错误处理
	jsonCodec := &JSONCodec{}
	
	// 测试编码nil envelope
	var buf bytes.Buffer
	err := jsonCodec.Encode(&buf, nil)
	if err == nil {
		t.Error("Expected error when encoding nil envelope with JSON codec")
	}
	if !containsSubstring(err.Error(), "JSONCodec.Encode") {
		t.Errorf("Error message should contain 'JSONCodec.Encode', got: %s", err.Error())
	}
	
	// 测试解码无效数据
	invalidData := bytes.NewReader([]byte("invalid json"))
	var envelope Envelope
	err = jsonCodec.Decode(invalidData, &envelope, 1024)
	if err == nil {
		t.Error("Expected error when decoding invalid JSON")
	}
	if !containsSubstring(err.Error(), "JSONCodec.Decode") {
		t.Errorf("Error message should contain 'JSONCodec.Decode', got: %s", err.Error())
	}
	
	// 测试Protobuf编码器错误处理
	protoCodec := &ProtobufCodec{}
	
	// 测试编码nil envelope
	buf.Reset()
	err = protoCodec.Encode(&buf, nil)
	if err == nil {
		t.Error("Expected error when encoding nil envelope with Protobuf codec")
	}
	if !containsSubstring(err.Error(), "ProtobufCodec.Encode") {
		t.Errorf("Error message should contain 'ProtobufCodec.Encode', got: %s", err.Error())
	}
	
	// 测试解码无效数据
	invalidProtoData := bytes.NewReader([]byte("invalid protobuf"))
	envelope = Envelope{}
	err = protoCodec.Decode(invalidProtoData, &envelope, 1024)
	if err == nil {
		t.Error("Expected error when decoding invalid Protobuf")
	}
	if !containsSubstring(err.Error(), "ProtobufCodec.Decode") {
		t.Errorf("Error message should contain 'ProtobufCodec.Decode', got: %s", err.Error())
	}
}

// containsSubstring 检查字符串是否包含子字符串
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 bytes.Contains([]byte(s), []byte(substr)))))
}