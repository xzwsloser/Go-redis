package handler

import (
	"context"
	"net"
)

type HandleFunc func(context.Context, net.Conn)

// Handler is the application method invoker
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
