package transport

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
	"github.com/hongjun500/chat-go/pkg/logger"
)

type TCPConn struct {
	conn net.Conn
	r    *bufio.Reader
}

type Options struct {
	OutBuffer int
	// TCP
	TCPMode      TCPMode
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	MaxFrameSize int // for JSON mode, defaults to 1MB
}

func NewTCPConn(c net.Conn) *TCPConn {
	return &TCPConn{conn: c, r: bufio.NewReader(c)}
}

func (t *TCPConn) ReadLine() (line string, err error) {
	line, err = t.r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (t *TCPConn) WriteLine(s string) error {
	_, err := fmt.Fprintln(t.conn, s)
	return err
}

func (t *TCPConn) Close() error { return t.conn.Close() }

// StartTcpWithRegistry 允许注入命令注册表，避免在 Hub 内部依赖命令，解开 import 循环
func StartTcpWithRegistry(addr string, hub *chat.Hub, reg *command.Registry) error {
	return StartTcpWithOptions(addr, hub, reg, Options{OutBuffer: 256, TCPMode: TCPModeJSON})
}

// StartTcpWithOptions 支持设置客户端发送缓冲区大小
func StartTcpWithOptions(addr string, hub *chat.Hub, reg *command.Registry, opt Options) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	logger.L().Sugar().Infow("tcp_listen", "addr", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.L().Sugar().Warnw("tcp_accept_error", "err", err)
			continue
		}
		go handleConnWithRegistryAndOptions(conn, hub, reg, opt)
	}
}

func handleConnWithRegistryAndOptions(conn net.Conn, hub *chat.Hub, reg *command.Registry, opt Options) {
	id := uuid.New().String()
	bufferSize := opt.OutBuffer
	if opt.MaxFrameSize <= 0 {
		opt.MaxFrameSize = 1 << 20 // 1MB default
	}
	if opt.TCPMode == "" {
		opt.TCPMode = TCPModeJSON
	}

	meta := make(map[string]string, 1)
	meta["level"] = "0"
	client := chat.NewClientWithBuffer(id, conn, bufferSize)
	client.Meta = meta

	switch opt.TCPMode {
	case TCPModeLegacy:
		// legacy line-based mode
		tc := NewTCPConn(conn)
		go func() {
			for msg := range client.Outgoing() {
				if opt.WriteTimeout > 0 {
					_ = conn.SetWriteDeadline(time.Now().Add(opt.WriteTimeout))
				}
				if err := tc.WriteLine(msg); err != nil {
					logger.L().Sugar().Warnw("tcp_write_error", "client", client.ID, "err", err)
					hub.UnregisterClient(client)
					return
				}
			}
		}()
		_ = tc.WriteLine("请输入昵称并回车：")
		reader := tc
		nameSet := false
		for {
			if opt.ReadTimeout > 0 {
				_ = conn.SetReadDeadline(time.Now().Add(opt.ReadTimeout))
			}
			line, err := reader.ReadLine()
			if err != nil {
				if err != io.EOF {
					logger.L().Sugar().Warnw("tcp_read_error", "client", client.ID, "err", err)
				}
				hub.UnregisterClient(client)
				return
			}
			if line == "" {
				continue
			}
			if !nameSet {
				if hub.IsBanned(line) {
					_ = tc.WriteLine("该用户已被封禁")
					hub.UnregisterClient(client)
					return
				}
				client.Name = line
				nameSet = true
				hub.RegisterClient(client)
				client.Send("昵称设置成功：" + line)
				continue
			}
			if strings.HasPrefix(line, "/") {
				handled, err := reg.Execute(line, &command.Context{Hub: hub, Client: client, Raw: line})
				if handled {
					if err != nil {
						client.Send("命令错误: " + err.Error())
					}
					continue
				}
			}
			hub.BroadcastLocal(client.Name, line)
			time.Sleep(1 * time.Millisecond)
		}
	default:
		// JSON framed mode
		codec := NewFrameCodec(conn)
		// writer drain
		go func() {
			for msg := range client.Outgoing() {
				if opt.WriteTimeout > 0 {
					_ = conn.SetWriteDeadline(time.Now().Add(opt.WriteTimeout))
				}
				_ = codec.Encode(&WireMessage{Type: "text", Text: msg, Ts: time.Now().UnixMilli()})
			}
		}()
		// greeting
		_ = codec.Encode(&WireMessage{Type: "text", Text: "请输入昵称并回车：", Ts: time.Now().UnixMilli()})
		nameSet := false
		for {
			if opt.ReadTimeout > 0 {
				_ = conn.SetReadDeadline(time.Now().Add(opt.ReadTimeout))
			}
			var m WireMessage
			if err := codec.Decode(&m, opt.MaxFrameSize); err != nil {
				if err != io.EOF {
					logger.L().Sugar().Warnw("tcp_decode_error", "client", client.ID, "err", err)
				}
				hub.UnregisterClient(client)
				return
			}
			switch m.Type {
			case "text":
				// 兼容：文本等同聊天内容/或首条为昵称
				if !nameSet {
					if hub.IsBanned(strings.TrimSpace(m.Text)) {
						_ = codec.Encode(&WireMessage{Type: "text", Text: "该用户已被封禁"})
						hub.UnregisterClient(client)
						return
					}
					client.Name = strings.TrimSpace(m.Text)
					nameSet = true
					hub.RegisterClient(client)
					_ = codec.Encode(&WireMessage{Type: "text", Text: "昵称设置成功：" + client.Name})
					continue
				}
				hub.BroadcastLocal(client.Name, m.Text)
			case "set_name":
				if nameSet {
					_ = codec.Encode(&WireMessage{Type: "ack", Status: "already_named"})
					continue
				}
				name := strings.TrimSpace(m.Name)
				if name == "" {
					_ = codec.Encode(&WireMessage{Type: "ack", Status: "invalid_name"})
					continue
				}
				if hub.IsBanned(name) {
					_ = codec.Encode(&WireMessage{Type: "text", Text: "该用户已被封禁"})
					hub.UnregisterClient(client)
					return
				}
				client.Name = name
				nameSet = true
				hub.RegisterClient(client)
				_ = codec.Encode(&WireMessage{Type: "ack", Status: "ok"})
			case "chat":
				if !nameSet {
					_ = codec.Encode(&WireMessage{Type: "ack", Status: "unauthorized"})
					continue
				}
				hub.BroadcastLocal(client.Name, m.Content)
			case "direct":
				if !nameSet {
					_ = codec.Encode(&WireMessage{Type: "ack", Status: "unauthorized"})
					continue
				}
				hub.Emit(&chat.DirectMessageEvent{When: time.Now(), From: client.Name, To: m.To, Content: m.Content})
			case "command":
				handled, err := reg.Execute(m.Raw, &command.Context{Hub: hub, Client: client, Raw: m.Raw})
				if handled && err != nil {
					_ = codec.Encode(&WireMessage{Type: "text", Text: "命令错误: " + err.Error()})
				}
			case "ping":
				_ = codec.Encode(&WireMessage{Type: "pong", Seq: m.Seq})
			default:
				_ = codec.Encode(&WireMessage{Type: "ack", Status: "unknown_type"})
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}
