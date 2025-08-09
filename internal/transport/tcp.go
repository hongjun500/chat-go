package transport

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"github.com/hongjun500/chat-go/internal/chat"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type TCPConn struct {
	conn net.Conn
	r    *bufio.Reader
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

func StartTcp(addr string, hub *chat.Hub) error {
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
		go handleConn(conn, hub)
	}
}

func handleConn(conn net.Conn, hub *chat.Hub) {
	id := uuid.New().String()
	tc := NewTCPConn(conn)
	client := chat.NewClient(id, conn)
	hub.RegisterClient(client)

	go func() {
		for msg := range client.Outgoing() {
			if err := tc.WriteLine(msg); err != nil {
				// 写失败（网络问题），注销客户端并退出写协程
				log.Printf("write err to %s: %v", client.ID, err)
				hub.UnregisterClient(client)
				return
			}
		}
	}()

	client.Send("请输入昵称并回车：")

	reader := tc
	nameSet := false
	for {
		line, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				hub.UnregisterClient(client)
				return
			}
			log.Printf("read from client err: %v", err)
			hub.UnregisterClient(client)
			return
		}
		if line == "" {
			continue
		}
		if !nameSet {
			client.Name = line
			nameSet = true
			client.Send("昵称设置成功：" + line)
			hub.BroadcastLocal(client.Name, "加入了聊天室")
			continue
		}
		// 命令模式
		// 简单命令支持：/who /help /quit （阶段1，命令后续会抽象）
		if strings.HasPrefix(line, "/") {
			switch line {
			case "/who":
				client.Send("在线: " + strings.Join(hub.ListNames(), ", "))
			case "/help":
				client.Send("/who 查看在线, /quit 退出, /help 帮助")
			case "/quit":
				client.Send("再见")
				hub.UnregisterClient(client)
				return
			default:
				client.Send("未知命令: " + line + " (支持 /who /help /quit )")
			}
			continue
		}

		// 普通消息 -> 触发本地事件（异步分发给订阅者）
		hub.BroadcastLocal(client.Name, line)

		// small safety sleep to avoid tight loop; optional
		time.Sleep(1 * time.Millisecond)
	}

}
