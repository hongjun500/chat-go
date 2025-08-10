package transport

import (
	"strings"
	"time"

	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
)

// GatewayHandler hosts auth/nickname, commands and chat routing in a protocol-agnostic way
type GatewayHandler struct {
	Hub      *chat.Hub
	Commands *command.Registry
}

func (g *GatewayHandler) OnSessionOpen(sess Session) {
	_ = sess.SendEnvelope(&Envelope{Type: "text", Text: "请输入昵称并回车：", Ts: time.Now().UnixMilli()})
}

func (g *GatewayHandler) OnEnvelope(sess Session, m *Envelope) {
	// find client by session.ID via Hub? Current Hub manages *chat.Client; we create one lazily on first name
	// We reuse prior implementation by storing mapping in Client.Meta; embed session ID into client ID.
	// For now, we locate by iterating or we carry the *chat.Client in session implementation. Simpler: each session impl holds *chat.Client.
	// Therefore this layer expects Session to wrap a *chat.Client and route via that.
	ts := time.Now().UnixMilli()
	switch m.Type {
	case "text":
		// treat as nickname if not set; else as chat
		c := getClient(sess)
		if c.Name == "" {
			name := strings.TrimSpace(m.Text)
			if name == "" {
				_ = sess.SendEnvelope(&Envelope{Type: "ack", Status: "invalid_name", Ts: ts})
				return
			}
			if g.Hub.IsBanned(name) {
				_ = sess.SendEnvelope(&Envelope{Type: "text", Text: "该用户已被封禁", Ts: ts})
				g.Hub.UnregisterClient(c)
				return
			}
			c.Name = name
			g.Hub.RegisterClient(c)
			_ = sess.SendEnvelope(&Envelope{Type: "text", Text: "昵称设置成功：" + c.Name, Ts: ts})
			return
		}
		g.Hub.BroadcastLocal(c.Name, m.Text)
	case "set_name":
		c := getClient(sess)
		if c.Name != "" {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Status: "already_named", Ts: ts})
			return
		}
		name := strings.TrimSpace(m.Name)
		if name == "" {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Status: "invalid_name", Ts: ts})
			return
		}
		if g.Hub.IsBanned(name) {
			_ = sess.SendEnvelope(&Envelope{Type: "text", Text: "该用户已被封禁", Ts: ts})
			g.Hub.UnregisterClient(c)
			return
		}
		c.Name = name
		g.Hub.RegisterClient(c)
		_ = sess.SendEnvelope(&Envelope{Type: "ack", Status: "ok", Ts: ts})
	case "chat":
		c := getClient(sess)
		if c.Name == "" {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Status: "unauthorized", Ts: ts})
			return
		}
		g.Hub.BroadcastLocal(c.Name, m.Content)
	case "direct":
		c := getClient(sess)
		if c.Name == "" {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Status: "unauthorized", Ts: ts})
			return
		}
		g.Hub.Emit(&chat.DirectMessageEvent{When: time.Now(), From: c.Name, To: m.To, Content: m.Content})
	case "command":
		c := getClient(sess)
		handled, err := g.Commands.Execute(m.Raw, &command.Context{Hub: g.Hub, Client: c, Raw: m.Raw})
		if handled && err != nil {
			_ = sess.SendEnvelope(&Envelope{Type: "text", Text: "命令错误: " + err.Error(), Ts: ts})
		}
	case "ping":
		_ = sess.SendEnvelope(&Envelope{Type: "pong", Seq: m.Seq, Ts: ts})
	default:
		_ = sess.SendEnvelope(&Envelope{Type: "ack", Status: "unknown_type", Ts: ts})
	}
}

func (g *GatewayHandler) OnSessionClose(sess Session, err error) {
	c := getClient(sess)
	if c != nil {
		g.Hub.UnregisterClient(c)
	}
}
