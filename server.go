package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var (
	ClientsMux sync.Mutex
	Broadcast  = make(chan string)
	Upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	Clients      = make(map[*Client]bool)
	ClientByName = make(map[string]*Client)
	Messages     = make(chan string)
)

func handleWS(w http.ResponseWriter, r *http.Request) {
	ws, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	_, msg, err := ws.ReadMessage()
	if err != nil {
		ws.Close()
		return
	}
	name := string(msg)
	cli := &Client{
		Conn: ws,
		Name: name,
	}
	ClientsMux.Lock()
	Clients[cli] = true
	ClientByName[name] = cli
	ClientsMux.Unlock()

	Broadcast <- fmt.Sprintf("%s connected and join", name)

	go handleClient(cli)
}

func handleClient(cli *Client) {
	defer func() {
		ClientsMux.Lock()
		delete(Clients, cli)
		ClientsMux.Unlock()
		cli.Conn.Close()
		Broadcast <- fmt.Sprintf("%s disconnected", cli.Name)
	}()

	for {
		_, msg, err := cli.Conn.ReadMessage()
		if err != nil {
			break
		}
		Broadcast <- fmt.Sprintf("[%s] %s", cli.Name, string(msg))
	}
}

func broadcaster() {
	for {
		msg := <-Broadcast
		ClientsMux.Lock()
		for c := range Clients {
			c.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
		ClientsMux.Unlock()
	}
}

// sendUserList 在线用户列表
/*func sendUserList(conn net.Conn) {
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
}*/
