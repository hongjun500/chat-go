package main

import (
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/transport"
)

func main() {
	hub := chat.NewHub()
	// 本地广播 handler：把 MessageEvent 转为文本并发送
	hub.Subscribe(chat.EventMessageLocal, func(e chat.Event) {
		me := e.(*chat.MessageEvent)
		text := "[" + me.When.Format("2006-01-02 15:01:05") + "] " + me.From + ": " + me.Content
		hub.SendToAll(text)
	})
	hub.Subscribe(chat.EventMessageRemote, func(e chat.Event) {
		me := e.(*chat.MessageEvent)
		text := "[" + me.When.Format("2006-01-02 15:01:05") + "] " + me.From + ": " + me.Content + " (来自远端)"
		hub.SendToAll(text)
	})

	// 用户上下线通知（可选）
	hub.Subscribe(chat.EventUserJoined, func(e chat.Event) {
		ue := e.(*chat.UserEvent)
		hub.SendToAll("[系统] " + ue.User.Name + " 加入")
	})
	hub.Subscribe(chat.EventUserLeave, func(e chat.Event) {
		ue := e.(*chat.UserEvent)
		hub.SendToAll("[系统] " + ue.User.Name + " 离开")
	})

	transport.StartTcp(":8080", hub)
}
