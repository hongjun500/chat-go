package transport

import (
	"fmt"
	"github.com/hongjun500/chat-go/internal/protocol"
	"time"
)

var hds = make(map[string]func(Session, *protocol.Envelope) error)

func init() {
	d := NewEnvelopeDispatcher()
	hds[string(protocol.MsgText)] = d.textHandler
}

type EnvelopeDispatcher struct {
	ptl      *protocol.Protocol
	handlers map[string]func(Session, *protocol.Envelope) error
}

func NewEnvelopeDispatcher() *EnvelopeDispatcher {
	d := &EnvelopeDispatcher{
		ptl:      protocol.DefaultProtocol,
		handlers: hds,
	}
	return d
}

func (d *EnvelopeDispatcher) Welcome(sess Session) {
	envelope, _ := d.ptl.Welcome("请输入昵称并回车：")
	envelope.Ts = time.Now().UnixMilli()
	_ = sess.SendEnvelope(envelope)
}

func (d *EnvelopeDispatcher) Dispatch(sess Session, e *protocol.Envelope) error {
	if handler, ok := d.handlers[string(e.Type)]; ok {
		return handler(sess, e)
	}
	// 如果没有找到对应的处理函数，可以选择返回错误或者忽略
	return fmt.Errorf("no handler for message type: %s", e.Type)
}

func (d *EnvelopeDispatcher) textHandler(sess Session, e *protocol.Envelope) error {
	/*c := getClient(sess)
	if c.Name == "" {
		err := d.ptl.Dispatch(e)
		//p, err := g.PayloadDecoder.DecodeText(m)
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
	g.Hub.BroadcastLocal(c.Name, p.Text)*/
	return nil
}
