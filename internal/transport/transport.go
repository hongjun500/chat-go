package transport

import (
	"context"
)

// Session represents a logical client connection independent of underlying protocol
type Session interface {
	ID() string
	RemoteAddr() string
	SendEnvelope(*Envelope) error
	Close() error
}

// Gateway consumes high-level Envelope from any Transport
type Gateway interface {
	OnSessionOpen(sess Session)
	OnEnvelope(sess Session, msg *Envelope)
	OnSessionClose(sess Session, err error)
}

// Transport runs a server endpoint and delivers messages to the gateway
type Transport interface {
	Name() string
	Start(ctx context.Context, addr string, gateway Gateway, opt Options) error
}
