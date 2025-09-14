package protocol

import (
	"testing"
)

// TestCodecConfig 测试编码器配置功能
func TestCodecConfig(t *testing.T) {
	config := NewDefaultCodecConfig(CodecJson)
	
	// 测试获取默认编码器
	if config.GetDefaultCodec() != CodecJson {
		t.Errorf("Expected default codec %d, got %d", CodecJson, config.GetDefaultCodec())
	}
	
	// 测试设置默认编码器
	err := config.SetDefaultCodec(CodecProtobuf)
	if err != nil {
		t.Errorf("Failed to set default codec: %v", err)
	}
	
	if config.GetDefaultCodec() != CodecProtobuf {
		t.Errorf("Expected default codec %d after setting, got %d", CodecProtobuf, config.GetDefaultCodec())
	}
	
	// 测试设置不支持的编码器
	err = config.SetDefaultCodec(999)
	if err == nil {
		t.Error("Expected error when setting unsupported codec")
	}
	
	// 测试根据名称获取编码器
	codecType, err := config.GetCodecByName(Json)
	if err != nil {
		t.Errorf("Failed to get codec by name: %v", err)
	}
	if codecType != CodecJson {
		t.Errorf("Expected codec type %d for name %s, got %d", CodecJson, Json, codecType)
	}
	
	// 测试不支持的编码器名称
	_, err = config.GetCodecByName("unsupported")
	if err == nil {
		t.Error("Expected error for unsupported codec name")
	}
	
	// 测试编码器支持检查
	if !config.IsCodecSupported(CodecJson) {
		t.Error("JSON codec should be supported")
	}
	
	if config.IsCodecSupported(999) {
		t.Error("Unsupported codec should not be reported as supported")
	}
}

// TestProtocolDynamicConfig 测试协议动态配置
func TestProtocolDynamicConfig(t *testing.T) {
	protocol := NewProtocol(CodecJson)
	
	// 测试获取当前编码器
	name, codecType := protocol.GetCurrentCodec()
	if name != Json || codecType != CodecJson {
		t.Errorf("Expected current codec %s/%d, got %s/%d", Json, CodecJson, name, codecType)
	}
	
	// 测试设置编码器
	err := protocol.SetCodec(CodecProtobuf)
	if err != nil {
		t.Errorf("Failed to set codec: %v", err)
	}
	
	name, codecType = protocol.GetCurrentCodec()
	if name != Protobuf || codecType != CodecProtobuf {
		t.Errorf("Expected current codec %s/%d after setting, got %s/%d", Protobuf, CodecProtobuf, name, codecType)
	}
	
	// 测试根据名称设置编码器
	err = protocol.SetCodecByName(Json)
	if err != nil {
		t.Errorf("Failed to set codec by name: %v", err)
	}
	
	name, codecType = protocol.GetCurrentCodec()
	if name != Json || codecType != CodecJson {
		t.Errorf("Expected current codec %s/%d after setting by name, got %s/%d", Json, CodecJson, name, codecType)
	}
	
	// 测试设置不支持的编码器
	err = protocol.SetCodec(999)
	if err == nil {
		t.Error("Expected error when setting unsupported codec")
	}
}

// TestNewProtocolWithConfig 测试使用配置创建协议
func TestNewProtocolWithConfig(t *testing.T) {
	config := NewDefaultCodecConfig(CodecProtobuf)
	
	protocol, err := NewProtocolWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create protocol with config: %v", err)
	}
	
	name, codecType := protocol.GetCurrentCodec()
	if name != Protobuf || codecType != CodecProtobuf {
		t.Errorf("Expected protocol codec %s/%d, got %s/%d", Protobuf, CodecProtobuf, name, codecType)
	}
	
	// 测试使用不支持的编码器配置
	badConfig := NewDefaultCodecConfig(CodecJson)
	// 直接修改配置的内部状态来创建无效配置
	badConfig.defaultCodec = 999
	_, err = NewProtocolWithConfig(badConfig)
	if err == nil {
		t.Error("Expected error when creating protocol with unsupported codec config")
	}
}