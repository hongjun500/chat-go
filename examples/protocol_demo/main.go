package main

import (
	"fmt"
	"log"

	"github.com/hongjun500/chat-go/internal/protocol"
)

// 演示新的协议管理器功能
func main() {
	fmt.Println("=== Chat-Go Protocol Manager Demo ===")

	// 1. 创建协议管理器，默认使用JSON编码器和版本1.0
	pm, err := protocol.NewProtocolManager(protocol.CodecJson, "1.0")
	if err != nil {
		log.Fatalf("创建协议管理器失败: %v", err)
	}

	fmt.Printf("初始编码器: %s\n", getCurrentCodecInfo(pm))

	// 2. 演示动态切换编码器
	fmt.Println("\n=== 动态编码器切换 ===")
	
	err = pm.SetDefaultCodecByName("protobuf")
	if err != nil {
		log.Printf("切换到Protobuf编码器失败: %v", err)
	} else {
		fmt.Printf("切换后编码器: %s\n", getCurrentCodecInfo(pm))
	}

	// 切换回JSON
	err = pm.SetDefaultCodec(protocol.CodecJson)
	if err != nil {
		log.Printf("切换到JSON编码器失败: %v", err)
	} else {
		fmt.Printf("切换回JSON编码器: %s\n", getCurrentCodecInfo(pm))
	}

	// 3. 演示消息创建和编解码
	fmt.Println("\n=== 消息创建和编解码 ===")
	
	// 创建文本消息
	textMsg, err := pm.CreateTextMessage("Hello, World!", "alice", "bob")
	if err != nil {
		log.Printf("创建文本消息失败: %v", err)
		return
	}
	fmt.Printf("创建的文本消息: Type=%s, From=%s, To=%s\n", textMsg.Type, textMsg.From, textMsg.To)

	// 编码消息
	data, err := pm.EncodeMessage(textMsg)
	if err != nil {
		log.Printf("编码消息失败: %v", err)
		return
	}
	fmt.Printf("编码后的消息大小: %d bytes\n", len(data))

	// 解码消息
	decoded, err := pm.DecodeMessage(data, 1024*1024)
	if err != nil {
		log.Printf("解码消息失败: %v", err)
		return
	}
	fmt.Printf("解码的消息: Type=%s, From=%s, To=%s\n", decoded.Type, decoded.From, decoded.To)

	// 4. 演示版本控制
	fmt.Println("\n=== 版本控制 ===")
	
	fmt.Printf("默认版本: %s\n", pm.GetDefaultVersion())
	fmt.Printf("支持的版本: %v\n", pm.GetSupportedVersions())

	// 注册自定义处理器
	customHandler := func(env *protocol.Envelope) error {
		fmt.Printf("自定义处理器处理消息: %s (Version=%s)\n", env.Type, env.Version)
		return nil
	}

	// 注册版本特定的处理器
	err = pm.RegisterVersionedHandler("1.1", protocol.MsgText, customHandler, protocol.CodecProtobuf)
	if err != nil {
		log.Printf("注册版本处理器失败: %v", err)
	} else {
		fmt.Printf("成功注册版本1.1的文本消息处理器\n")
		fmt.Printf("更新后支持的版本: %v\n", pm.GetSupportedVersions())
	}

	// 5. 演示错误处理改进
	fmt.Println("\n=== 增强的错误处理 ===")
	
	// 尝试解码无效数据
	invalidData := []byte("invalid json data")
	_, err = pm.DecodeMessage(invalidData, 1024)
	if err != nil {
		fmt.Printf("解码无效数据的错误信息: %v\n", err)
	}

	// 尝试编码空消息
	_, err = pm.EncodeMessage(nil)
	if err != nil {
		fmt.Printf("编码空消息的错误信息: %v\n", err)
	}

	// 6. 演示不同消息类型
	fmt.Println("\n=== 不同消息类型 ===")
	
	// 创建聊天消息
	chatMsg, err := pm.CreateChatMessage("这是一条聊天消息", "alice")
	if err == nil {
		fmt.Printf("聊天消息: Type=%s, From=%s\n", chatMsg.Type, chatMsg.From)
	}

	// 创建确认消息
	ackMsg, err := pm.CreateAckMessage("ok", "test-correlation-id")
	if err == nil {
		fmt.Printf("确认消息: Type=%s, Correlation=%s\n", ackMsg.Type, ackMsg.Correlation)
	}

	// 创建命令消息
	cmdMsg, err := pm.CreateCommandMessage("/help", "alice")
	if err == nil {
		fmt.Printf("命令消息: Type=%s, From=%s\n", cmdMsg.Type, cmdMsg.From)
	}

	fmt.Println("\n=== Demo 完成 ===")
}

// 获取当前编码器信息
func getCurrentCodecInfo(pm *protocol.ProtocolManager) string {
	name, codecType := pm.GetCurrentCodec()
	return fmt.Sprintf("%s (type=%d)", name, codecType)
}