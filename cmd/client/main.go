package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hongjun500/chat-go/internal/codec"
	"github.com/hongjun500/chat-go/internal/protocol"
	"github.com/hongjun500/chat-go/internal/transport"
)

func main() {
	var (
		addr      = flag.String("addr", "localhost:8080", "服务器地址")
		codecType = flag.String("codec", "json", "编码类型 (json/protobuf)")
		username  = flag.String("user", "testuser", "用户名")
	)
	flag.Parse()

	fmt.Printf("连接到 %s，使用 %s 编码...\n", *addr, *codecType)

	// 创建编解码器
	messageCodec, err := codec.NewCodec(*codecType)
	if err != nil {
		log.Fatalf("创建编解码器失败: %v", err)
	}

	// 连接到 TCP 服务器
	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	fmt.Printf("已连接到 %s，内容类型: %s\n", *addr, messageCodec.ContentType())

	// 创建帧编解码器
	frameCodec := transport.NewFrameCodec(conn)
	framedCodec := transport.NewFramedMessageCodec(frameCodec, messageCodec)

	// 发送设置昵称消息
	setNamePayload, _ := json.Marshal(protocol.SetNamePayload{Name: *username})
	setNameMsg := &protocol.Envelope{
		Type:    "set_name",
		Mid:     "set-name-001",
		Ts:      time.Now().UnixMilli(),
		Payload: json.RawMessage(setNamePayload),
	}

	fmt.Printf("发送设置昵称消息: %s\n", *username)
	if err := framedCodec.Encode(setNameMsg); err != nil {
		log.Fatalf("发送设置昵称消息失败: %v", err)
	}

	// 读取响应
	var ackMsg protocol.Envelope
	if err := framedCodec.Decode(&ackMsg, 1024*1024); err != nil {
		log.Fatalf("读取响应失败: %v", err)
	}

	fmt.Printf("收到响应: Type=%s, Mid=%s\n", ackMsg.Type, ackMsg.Mid)
	if ackMsg.Type == "ack" {
		var ackPayload protocol.AckPayload
		if err := json.Unmarshal(ackMsg.Payload, &ackPayload); err == nil {
			fmt.Printf("设置昵称状态: %s\n", ackPayload.Status)
		}
	}

	// 发送聊天消息
	chatPayload, _ := json.Marshal(protocol.ChatPayload{Content: fmt.Sprintf("Hello from %s client!", *codecType)})
	chatMsg := &protocol.Envelope{
		Type:    "chat",
		Mid:     "chat-001",
		From:    *username,
		Ts:      time.Now().UnixMilli(),
		Payload: json.RawMessage(chatPayload),
	}

	fmt.Printf("发送聊天消息...\n")
	if err := framedCodec.Encode(chatMsg); err != nil {
		log.Fatalf("发送聊天消息失败: %v", err)
	}

	// 发送ping消息测试心跳
	pingPayload, _ := json.Marshal(protocol.PingPayload{Seq: 1, Timestamp: time.Now().UnixMilli()})
	pingMsg := &protocol.Envelope{
		Type:    protocol.MsgPing,
		Mid:     "ping-001",
		Ts:      time.Now().UnixMilli(),
		Payload: json.RawMessage(pingPayload),
	}

	fmt.Printf("发送 ping 消息...\n")
	if err := framedCodec.Encode(pingMsg); err != nil {
		log.Fatalf("发送 ping 消息失败: %v", err)
	}

	// 读取 pong 响应
	var pongMsg protocol.Envelope
	if err := framedCodec.Decode(&pongMsg, 1024*1024); err != nil {
		log.Fatalf("读取 pong 响应失败: %v", err)
	}

	fmt.Printf("收到 pong 响应: Type=%s\n", pongMsg.Type)
	if pongMsg.Type == protocol.MsgPong {
		var pongPayload protocol.PongPayload
		if err := json.Unmarshal(pongMsg.Payload, &pongPayload); err == nil {
			fmt.Printf("Pong Seq: %d\n", pongPayload.Seq)
		}
	}

	// 发送命令消息
	cmdPayload, _ := json.Marshal(protocol.CommandPayload{Raw: "/help"})
	cmdMsg := &protocol.Envelope{
		Type:    protocol.MsgCommand,
		Mid:     "cmd-001",
		Ts:      time.Now().UnixMilli(),
		Payload: json.RawMessage(cmdPayload),
	}

	fmt.Printf("发送命令消息: /help\n")
	if err := framedCodec.Encode(cmdMsg); err != nil {
		log.Fatalf("发送命令消息失败: %v", err)
	}

	// 读取更多消息（最多 3 秒）
	fmt.Printf("监听更多消息 (3秒)...\n")
	
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		var msg protocol.Envelope
		if err := framedCodec.Decode(&msg, 1024*1024); err != nil {
			fmt.Printf("读取超时或连接关闭: %v\n", err)
			break
		}

		fmt.Printf("收到消息: Type=%s, From=%s\n", msg.Type, msg.From)
		if msg.Type == protocol.MsgText {
			var textPayload protocol.TextPayload
			if err := json.Unmarshal(msg.Payload, &textPayload); err == nil {
				fmt.Printf("  内容: %s\n", textPayload.Text)
			}
		}
	}

	fmt.Printf("客户端演示完成\n")
}