package transport

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
)

type TCPConn struct {
	conn net.Conn
	r    *bufio.Reader
}

type Options struct {
	OutBuffer int
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
	return StartTcpWithOptions(addr, hub, reg, Options{OutBuffer: 256})
}

// StartTcpWithOptions 支持设置客户端发送缓冲区大小
func StartTcpWithOptions(addr string, hub *chat.Hub, reg *command.Registry, opt Options) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("TCP Listen on %s", addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("tcp accept err:%v", err)
			continue
		}
		go handleConnWithRegistryAndOptions(conn, hub, reg, opt)
	}
}

func handleConnWithRegistryAndOptions(conn net.Conn, hub *chat.Hub, reg *command.Registry, opt Options) {
	id := uuid.New().String()
	tc := NewTCPConn(conn)
	bufferSize := opt.OutBuffer
	meta := make(map[string]string, 1)
	meta["level"] = "0"
	client := chat.NewClientWithBuffer(id, conn, bufferSize)
	client.Meta = meta
	// 写出协程
	go func() {
		for msg := range client.Outgoing() {
			if err := tc.WriteLine(msg); err != nil {
				log.Printf("write err to %s: %v", client.ID, err)
				hub.UnregisterClient(client)
				return
			}
		}
	}()

	_ = tc.WriteLine("请输入昵称并回车：")

	reader := tc
	nameSet := false
	for {
		line, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Printf("read from client err: %v", err)
			}
			hub.UnregisterClient(client)
			return
		}
		if line == "" {
			continue
		}
		if !nameSet {
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
}
