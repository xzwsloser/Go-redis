package handler

import (
	"bufio"
	"context"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/lib/sync/atomic"
	"github.com/xzwsloser/Go-redis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

var (
	MaxTimeout time.Duration = 10 * time.Second
)

type EchoHandler struct {
	activeConn sync.Map
	isClosed   atomic.Boolean
}

type EchoClient struct {
	Conn net.Conn
	wait wait.Wait
}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (e *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if e.isClosed.Get() {
		_ = conn.Close()
		return
	}

	client := &EchoClient{
		Conn: conn,
	}

	e.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				e.activeConn.Delete(client)
				_ = client.Close()
				continue
			}

			logger.Error("echo handler err: ", err.Error())
			return
		}

		client.wait.Add(1)
		_, _ = conn.Write([]byte(msg))
		client.wait.Done()
	}
}

func (e *EchoHandler) Close() error {
	e.isClosed.Set(true)
	e.activeConn.Range(func(k, v any) bool {
		_ = k.(*EchoClient).Close()
		return true
	})
	return nil
}

func (ec *EchoClient) Close() error {
	ec.wait.WaitWithTimeout(MaxTimeout)
	_ = ec.Conn.Close()
	return nil
}
