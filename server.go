package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	Clients      = make(map[net.Conn]*Client)
	ClientByName = make(map[string]*Client)
	ClientsMux   sync.Mutex
	Messages     = make(chan string)
)

// handleConnection 连接处理
func handleConnection(conn net.Conn) {
	defer func() {
		ClientsMux.Lock()
		client := Clients[conn]
		delete(Clients, conn)
		delete(ClientByName, client.Name)
		ClientsMux.Unlock()
		Messages <- fmt.Sprintf("🛑 %s disconnected", client.Name)
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

	ClientsMux.Lock()
	c := &Client{
		Conn: conn,
		Name: name,
	}
	Clients[conn] = c
	ClientByName[name] = c
	ClientsMux.Unlock()

	Messages <- fmt.Sprintf("%s join chat-go", name)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "/list" {
			sendUserList(conn)
			continue
		}
		if strings.HasPrefix(text, "@") && strings.Contains(text, ":") {
			handlePrivateMessage(conn, text)
			continue
		}
		msg := fmt.Sprintf("[%s]: %s", name, text)
		Messages <- msg
	}
}

// sendUserList 在线用户列表
func sendUserList(conn net.Conn) {
	ClientsMux.Lock()
	defer ClientsMux.Unlock()

	conn.Write([]byte("List of online users:\n"))

	for _, cli := range Clients {
		conn.Write([]byte("- " + cli.Name + "\n"))
	}
}

// broadcast 广播通知
func broadcast() {
	for msg := range Messages {
		ClientsMux.Lock()
		for _, cli := range Clients {
			fmt.Fprintln(cli.Conn, msg)
		}
		ClientsMux.Unlock()
	}
}

// handlePrivateMessage 私聊
func handlePrivateMessage(conn net.Conn, text string) {
	sender := Clients[conn]

	parts := strings.SplitN(text[1:], ":", 2)
	if len(parts) != 2 {
		conn.Write([]byte("❌private msg error, should @name:content"))
		return
	}
	targetName := parts[0]
	msg := parts[1]
	ClientsMux.Lock()
	target, ok := ClientByName[targetName]
	ClientsMux.Unlock()
	if ok {
		// 给目标发送消息
		fmt.Fprintf(target.Conn, "💌private msg from [%s]: %s\n", sender.Name, msg)
		// 也给自己确认一下
		fmt.Fprintf(sender.Conn, "📤private msg sended [%s]: %s\n", target.Name, msg)
	} else {
		conn.Write([]byte("❌user offline \n"))
	}
}
