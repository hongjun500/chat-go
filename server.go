package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	clients    = make(map[net.Conn]*Client)
	clientsMux sync.Mutex
	messages   = make(chan string)
)

// handleConnection 连接处理
func handleConnection(conn net.Conn) {
	defer func() {
		clientsMux.Lock()
		client := clients[conn]
		delete(clients, conn)
		clientsMux.Unlock()
		messages <- fmt.Sprintf("🛑 %s disconnected", client.Name)
		err := conn.Close()
		if err != nil {
			return
		}
	}()

	_, err := conn.Write([]byte("please input your name: "))
	if err != nil {
		fmt.Println(err)
	}
	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading name: ", err)
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Anonymous"
	}

	clientsMux.Lock()
	clients[conn] = &Client{
		Conn: conn,
		Name: name,
	}
	clientsMux.Unlock()

	messages <- fmt.Sprintf("%s join chat-go", name)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "/list" {
			sendUserList(conn)
			continue
		}
		msg := fmt.Sprintf("[%s]: %s", name, text)
		messages <- msg
	}
}

// sendUserList 在线用户列表
func sendUserList(conn net.Conn) {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	conn.Write([]byte("List of online users:\n"))

	for _, cli := range clients {
		conn.Write([]byte("- " + cli.Name + "\n"))
	}
}

// broadcast 广播通知
func broadcast() {
	for msg := range messages {
		clientsMux.Lock()
		for _, cli := range clients {
			fmt.Fprintln(cli.Conn, msg)
		}
		clientsMux.Unlock()
	}
}
