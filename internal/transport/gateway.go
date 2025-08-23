package transport

import (
	"encoding/json"
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
	// 提示信息通过 text 类型的 payload 传递
	payload, _ := json.Marshal(TextPayload{Text: "请输入昵称并回车："})
	_ = sess.SendEnvelope(&Envelope{Type: "text", Payload: payload, Ts: time.Now().UnixMilli()})
}

func (g *GatewayHandler) OnEnvelope(sess Session, m *Envelope) {
	// Session 实现中持有 *chat.Client，路由依赖该客户端的 Name 状态。
	ts := time.Now().UnixMilli()
	switch m.Type {
	case "text":
		// 如果还未设置昵称，则把 text 作为昵称；否则作为普通聊天消息
		c := getClient(sess)
		if c.Name == "" {
			var p TextPayload
			if err := json.Unmarshal(m.Payload, &p); err != nil {
				_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "bad_payload"}), Ts: ts})
				return
			}
			name := strings.TrimSpace(p.Text)
			if name == "" {
				_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "invalid_name"}), Ts: ts})
				return
			}
			if g.Hub.IsBanned(name) {
				_ = sess.SendEnvelope(&Envelope{Type: "text", Payload: mustJSON(TextPayload{Text: "该用户已被封禁"}), Ts: ts})
				g.Hub.UnregisterClient(c)
				return
			}
			c.Name = name
			g.Hub.RegisterClient(c)
			_ = sess.SendEnvelope(&Envelope{Type: "text", Payload: mustJSON(TextPayload{Text: "昵称设置成功：" + c.Name}), Ts: ts})
			return
		}
		var p TextPayload
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "bad_payload"}), Ts: ts})
			return
		}
		g.Hub.BroadcastLocal(c.Name, p.Text)
	case "set_name":
		c := getClient(sess)
		if c.Name != "" {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "already_named"}), Ts: ts})
			return
		}
		var p SetNamePayload
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "bad_payload"}), Ts: ts})
			return
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "invalid_name"}), Ts: ts})
			return
		}
		if g.Hub.IsBanned(name) {
			_ = sess.SendEnvelope(&Envelope{Type: "text", Payload: mustJSON(TextPayload{Text: "该用户已被封禁"}), Ts: ts})
			g.Hub.UnregisterClient(c)
			return
		}
		c.Name = name
		g.Hub.RegisterClient(c)
		_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "ok"}), Ts: ts})
	case "chat":
		c := getClient(sess)
		if c.Name == "" {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "unauthorized"}), Ts: ts})
			return
		}
		var p ChatPayload
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "bad_payload"}), Ts: ts})
			return
		}
		g.Hub.BroadcastLocal(c.Name, p.Content)
	case "direct":
		c := getClient(sess)
		if c.Name == "" {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "unauthorized"}), Ts: ts})
			return
		}
		var p DirectPayload
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "bad_payload"}), Ts: ts})
			return
		}
		// 保持现有 Direct 事件
		to := ""
		if len(p.To) > 0 {
			to = p.To[0]
		}
		g.Hub.Emit(&chat.DirectMessageEvent{When: time.Now(), From: c.Name, To: to, Content: p.Content})
	case "command":
		c := getClient(sess)
		var p CommandPayload
		if err := json.Unmarshal(m.Payload, &p); err != nil {
			_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "bad_payload"}), Ts: ts})
			return
		}
		handled, err := g.Commands.Execute(p.Raw, &command.Context{Hub: g.Hub, Client: c, Raw: p.Raw})
		if handled && err != nil {
			_ = sess.SendEnvelope(&Envelope{Type: "text", Payload: mustJSON(TextPayload{Text: "命令错误: " + err.Error()}), Ts: ts})
		}
	case "ping":
		var p PingPayload
		_ = json.Unmarshal(m.Payload, &p)
		_ = sess.SendEnvelope(&Envelope{Type: "pong", Payload: mustJSON(PongPayload{Seq: p.Seq}), Ts: ts})
	default:
		_ = sess.SendEnvelope(&Envelope{Type: "ack", Mid: m.Mid, Payload: mustJSON(AckPayload{Status: "unknown_type"}), Ts: ts})
	}
}

func (g *GatewayHandler) OnSessionClose(sess Session, err error) {
	c := getClient(sess)
	if c != nil {
		g.Hub.UnregisterClient(c)
	}
}

// getClient retrieves the chat.Client from various Session implementations
func getClient(sess Session) *chat.Client {
	// Try TCP session first
	if ts, ok := sess.(*tcpConn); ok {
		return ts.client
	}
	// Try WebSocket session
	if ws, ok := sess.(*wsConn); ok {
		return ws.client
	}
	return nil
}

// ---- Typed payloads and helpers ----

type TextPayload struct {
	Text string `json:"text"`
}
type SetNamePayload struct {
	Name string `json:"name"`
}
type ChatPayload struct {
	Content string `json:"content"`
}
type DirectPayload struct {
	Content string   `json:"content"`
	To      []string `json:"to"`
}
type CommandPayload struct {
	Raw string `json:"raw"`
}
type AckPayload struct {
	Status string `json:"status"`
}
type PingPayload struct {
	Seq int64 `json:"seq"`
}
type PongPayload struct {
	Seq int64 `json:"seq"`
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
