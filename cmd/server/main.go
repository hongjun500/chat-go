package main

import (
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
	"github.com/hongjun500/chat-go/internal/config"
	"github.com/hongjun500/chat-go/internal/transport"
)

func main() {
	cfg := config.Load()
	hub := chat.NewHub()
	// 初始化命令注册表（解环：在 main 中创建并传递）
	cmdReg := command.NewRegistry()
	if err := command.RegisterBuiltins(cmdReg); err != nil {
		panic(err)
	}
	// 本地广播 handler：把 MessageEvent 转为文本并发送
	hub.Subscribe(chat.EventMessageLocal, func(e chat.Event) {
		me := e.(*chat.MessageEvent)
		text := "[" + me.When.Format("2006-01-02 15:04:05") + "] " + me.From + ": " + me.Content
		hub.SendToAll(text)
	})
	hub.Subscribe(chat.EventMessageRemote, func(e chat.Event) {
		me := e.(*chat.MessageEvent)
		text := "[" + me.When.Format("2006-01-02 15:04:05") + "] " + me.From + ": " + me.Content + " (来自远端)"
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

	_ = transport.StartTcpWithRegistry(cfg.TCPAddr, hub, cmdReg)
	//_ = transport.StartTcpWithOptions(cfg.TCPAddr, hub, cmdReg, transport.Options{OutBuffer: cfg.OutBuffer})

}
