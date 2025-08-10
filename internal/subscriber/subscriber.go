package subscriber

import (
	"time"

	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/observe"
)

// RegisterAll 把所有内置订阅者注册到 Hub。业务可按需拆分不同订阅集。
func RegisterAll(hub *chat.Hub) {
	registerMessage(hub)
	registerUserLifecycle(hub)
	registerSystem(hub)
	registerFile(hub)
	registerHeartbeat(hub)
	registerDirect(hub)
}

func registerMessage(hub *chat.Hub) {
	hub.Subscribe(chat.EventMessageLocal, func(e chat.Event) {
		me := e.(*chat.MessageEvent)
		text := "[" + me.When.Format("2006-01-02 15:04:05") + "] " + me.From + ": " + me.Content
		hub.SendToAll(text)
		observe.IncMessage("local")
	})
	hub.Subscribe(chat.EventMessageRemote, func(e chat.Event) {
		// 预留：远端消息可由分布式总线消费端发出
		_ = e
		observe.IncMessage("remote")
	})
}

func registerUserLifecycle(hub *chat.Hub) {
	hub.Subscribe(chat.EventUserJoined, func(e chat.Event) {
		ue := e.(*chat.UserEvent)
		hub.SendToAll("[系统] " + ue.User.Name + " 加入")
	})
	hub.Subscribe(chat.EventUserLeave, func(e chat.Event) {
		ue := e.(*chat.UserEvent)
		hub.SendToAll("[系统] " + ue.User.Name + " 离开")
	})
}

func registerSystem(hub *chat.Hub) {
	hub.Subscribe(chat.EventSystemNotice, func(e chat.Event) {
		se := e.(*chat.SystemNoticeEvent)
		hub.SendToAll("[系统通知][" + se.Level + "] " + se.Content)
	})
}

func registerFile(hub *chat.Hub) {
	hub.Subscribe(chat.EventFileTransfer, func(e chat.Event) {
		fe := e.(*chat.FileTransferEvent)
		target := fe.To
		if target == "" || target == "*" {
			hub.SendToAll("[文件] " + fe.From + " -> 所有人: " + fe.FileName)
			return
		}
		hub.SendToAll("[文件] " + fe.From + " -> " + target + ": " + fe.FileName)
	})
}

func registerHeartbeat(hub *chat.Hub) {
	hub.Subscribe(chat.EventHeartbeat, func(e chat.Event) {
		// TODO: 这里可以记录 metrics 或最近心跳时间
		_ = time.Now()
		_ = e
		observe.IncHeartbeat()
	})
}

func registerDirect(hub *chat.Hub) {
	hub.Subscribe(chat.EventMessageDirect, func(e chat.Event) {
		de := e.(*chat.DirectMessageEvent)
		// TCP 客户端走 Hub 点对点
		sent := hub.SendToUser(de.To, "[私信] "+de.From+": "+de.Content)
		if !sent {
			// 找不到目标，可回执给发送者（也会被 WS 的连接侧订阅到，但这条主要面向 TCP）
			hub.SendToUser(de.From, "[系统] 用户不在线或不存在: "+de.To)
		}
		observe.IncDirect()
	})
}
