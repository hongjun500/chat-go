package transport

import (
	"github.com/hongjun500/chat-go/internal/chat"
	"github.com/hongjun500/chat-go/internal/command"
	"github.com/hongjun500/chat-go/internal/protocol"
)

type HandlerFunc func(sess Session, msg *protocol.Envelope)

var handlers = map[string]HandlerFunc{
	"text": nil,
}

// GatewayHandler hosts auth/nickname, commands and chat routing in a protocol-agnostic way
type GatewayHandler struct {
	Hub        *chat.Hub
	Commands   *command.Registry
	dispatcher *EnvelopeDispatcher
}

// NewGatewayHandler creates a new gateway handler with default payload codec
func NewGatewayHandler(hub *chat.Hub, commands *command.Registry) *GatewayHandler {
	g := &GatewayHandler{
		Hub:      hub,
		Commands: commands,
	}
	g.dispatcher = NewEnvelopeDispatcher()
	g.dispatcher.ptl = protocol.NewProtocol(protocol.CodecJson)
	return g
}

func (g *GatewayHandler) OnSessionOpen(sess Session) {
	g.dispatcher.Welcome(sess)
}

func (g *GatewayHandler) OnEnvelope(sess Session, m *protocol.Envelope) {
	/*err := g.dispatcher.Dispatch(sess, m)
	// Session 实现中持有 *chat.Client，路由依赖该客户端的 Name 状态。
	ts := time.Now().UnixMilli()
	switch m.Type {
	case "text":
		// 如果还未设置昵称，则把 text 作为昵称；否则作为普通聊天消息
		c := getClient(sess)
		if c.Name == "" {
			p, err := g.PayloadDecoder.DecodeText(m)
			if err != nil {
				ackEnv, _ := g.PayloadEncoder.EncodeAck("bad_payload")
				ackEnv.Mid = m.Mid
				ackEnv.Ts = ts
				_ = sess.SendEnvelope(ackEnv)
				return
			}
			name := strings.TrimSpace(p.Text)
			if name == "" {
				ackEnv, _ := g.PayloadEncoder.EncodeAck("invalid_name")
				ackEnv.Mid = m.Mid
				ackEnv.Ts = ts
				_ = sess.SendEnvelope(ackEnv)
				return
			}
			if g.Hub.IsBanned(name) {
				textEnv, _ := g.PayloadEncoder.EncodeText("该用户已被封禁")
				textEnv.Ts = ts
				_ = sess.SendEnvelope(textEnv)
				g.Hub.UnregisterClient(c)
				return
			}
			c.Name = name
			g.Hub.RegisterClient(c)
			textEnv, _ := g.PayloadEncoder.EncodeText("昵称设置成功：" + c.Name)
			textEnv.Ts = ts
			_ = sess.SendEnvelope(textEnv)
			return
		}
		p, err := g.PayloadDecoder.DecodeText(m)
		if err != nil {
			ackEnv, _ := g.PayloadEncoder.EncodeAck("bad_payload")
			ackEnv.Mid = m.Mid
			ackEnv.Ts = ts
			_ = sess.SendEnvelope(ackEnv)
			return
		}
		g.Hub.BroadcastLocal(c.Name, p.Text)
	case "set_name":
		c := getClient(sess)
		if c.Name != "" {
			ackEnv, _ := g.PayloadEncoder.EncodeAck("already_named")
			ackEnv.Mid = m.Mid
			ackEnv.Ts = ts
			_ = sess.SendEnvelope(ackEnv)
			return
		}
		p, err := g.PayloadDecoder.DecodeSetName(m)
		if err != nil {
			ackEnv, _ := g.PayloadEncoder.EncodeAck("bad_payload")
			ackEnv.Mid = m.Mid
			ackEnv.Ts = ts
			_ = sess.SendEnvelope(ackEnv)
			return
		}
		name := strings.TrimSpace(p.Name)
		if name == "" {
			ackEnv, _ := g.PayloadEncoder.EncodeAck("invalid_name")
			ackEnv.Mid = m.Mid
			ackEnv.Ts = ts
			_ = sess.SendEnvelope(ackEnv)
			return
		}
		if g.Hub.IsBanned(name) {
			textEnv, _ := g.PayloadEncoder.EncodeText("该用户已被封禁")
			textEnv.Ts = ts
			_ = sess.SendEnvelope(textEnv)
			g.Hub.UnregisterClient(c)
			return
		}
		c.Name = name
		g.Hub.RegisterClient(c)
		ackEnv, _ := g.PayloadEncoder.EncodeAck("ok")
		ackEnv.Mid = m.Mid
		ackEnv.Ts = ts
		_ = sess.SendEnvelope(ackEnv)
	case "chat":
		c := getClient(sess)
		if c.Name == "" {
			ackEnv, _ := g.PayloadEncoder.EncodeAck("unauthorized")
			ackEnv.Mid = m.Mid
			ackEnv.Ts = ts
			_ = sess.SendEnvelope(ackEnv)
			return
		}
		p, err := g.PayloadDecoder.DecodeChat(m)
		if err != nil {
			ackEnv, _ := g.PayloadEncoder.EncodeAck("bad_payload")
			ackEnv.Mid = m.Mid
			ackEnv.Ts = ts
			_ = sess.SendEnvelope(ackEnv)
			return
		}
		g.Hub.BroadcastLocal(c.Name, p.Content)
	case "direct":
		c := getClient(sess)
		if c.Name == "" {
			ackEnv, _ := g.PayloadEncoder.EncodeAck("unauthorized")
			ackEnv.Mid = m.Mid
			ackEnv.Ts = ts
			_ = sess.SendEnvelope(ackEnv)
			return
		}
		p, err := g.PayloadDecoder.DecodeDirect(m)
		if err != nil {
			ackEnv, _ := g.PayloadEncoder.EncodeAck("bad_payload")
			ackEnv.Mid = m.Mid
			ackEnv.Ts = ts
			_ = sess.SendEnvelope(ackEnv)
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
		p, err := g.PayloadDecoder.DecodeCommand(m)
		if err != nil {
			ackEnv, _ := g.PayloadEncoder.EncodeAck("bad_payload")
			ackEnv.Mid = m.Mid
			ackEnv.Ts = ts
			_ = sess.SendEnvelope(ackEnv)
			return
		}
		handled, err := g.Commands.Execute(p.Raw, &command.Context{Hub: g.Hub, Client: c, Raw: p.Raw})
		if handled && err != nil {
			textEnv, _ := g.PayloadEncoder.EncodeText("命令错误: " + err.Error())
			textEnv.Ts = ts
			_ = sess.SendEnvelope(textEnv)
		}
	case "ping":
		p, _ := g.PayloadDecoder.DecodePing(m)
		pongEnv, _ := g.PayloadEncoder.EncodePong(p.Seq)
		pongEnv.Ts = ts
		_ = sess.SendEnvelope(pongEnv)
	default:
		ackEnv, _ := g.PayloadEncoder.EncodeAck("unknown_type")
		ackEnv.Mid = m.Mid
		ackEnv.Ts = ts
		_ = sess.SendEnvelope(ackEnv)
	}*/
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
	// [TODO] ws
	// if ws, ok := sess.(*wsConn); ok {
	// return ws.client
	// }
	return nil
}
