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
	client := chat.NewClient(id, conn)
	hub.Register(client)

	go func() {
		for msg := range client.Outgoing() {
			_, err := fmt.Fprintln(conn, msg)
			if err != nil {
				log.Printf("write to client err: %v", err)
				break
			}
		}
	}()

	// 读循环
	reader := bufio.NewReader(conn)
	nameSet := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				hub.Unregister(client)
				conn.Close()
				return
			}
			log.Printf("read from client err: %v", err)
			hub.Unregister(client)
			conn.Close()
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !nameSet {
			client.Name = line
			nameSet = true
			client.Send("昵称设置成功：" + line)
			hub.Broadcast(&chat.Message{
				From:      client.Name,
				Content:   client.Name + "加入了聊天室",
				Timestamp: time.Now(),
			})
			continue
		}
		// 命令模式
		if strings.HasPrefix(line, "/") {
			handled := chat.ProcessCommandWrapper(hub, client, line)
			if handled {
				continue
			}
		}

		hub.Broadcast(&chat.Message{
			From:      client.Name,
			Content:   line,
			Timestamp: time.Now(),
		})
	}

}
