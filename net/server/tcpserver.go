package server

import (
	"context"
	"github.com/xzwsloser/Go-redis/config"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/net/handler"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type TcpServer struct {
	ip          string
	handler     handler.Handler
	closeCh     chan struct{}
	sigCh       chan os.Signal
	listener    net.Listener
	wait        sync.WaitGroup
	clientCount int32
	ctx         context.Context
	isClosed    bool
}

func NewTcpServer() *TcpServer {
	address := config.GetRedisServerConfig().Address
	port := config.GetRedisServerConfig().Port
	server := &TcpServer{
		ip:          address + ":" + strconv.Itoa(port),
		handler:     handler.NewEchoHandler(),
		closeCh:     make(chan struct{}),
		sigCh:       make(chan os.Signal, 1),
		clientCount: 0,
		ctx:         context.Background(),
		isClosed:    true,
	}

	signal.Notify(server.sigCh,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		switch <-server.sigCh {
		case syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT:
			server.closeCh <- struct{}{}
		default:
		}
	}()

	listener, err := net.Listen("tcp", server.ip)
	if err != nil {
		logger.Fatal("error: ", err.Error())
	}

	server.listener = listener
	return server
}

func (s *TcpServer) Run() {
	s.isClosed = false
	errCh := make(chan error, 1)
	defer close(errCh)
	go func() {
		select {
		case <-s.closeCh:
			logger.Info("tcp server is closing ...")
		case err := <-errCh:
			logger.Error("tcp server error: ", err.Error())
		}

		_ = s.listener.Close()
		_ = s.handler.Close()
	}()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// timeout try again
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				logger.Error("time out error: ", ne.Error())
				time.Sleep(5 * time.Millisecond)
				continue
			}
			errCh <- err
			break
		}

		s.wait.Add(1)
		atomic.AddInt32(&s.clientCount, 1)
		go func() {
			defer func() {
				s.wait.Done()
				atomic.AddInt32(&s.clientCount, -1)
			}()
			s.handler.Handle(s.ctx, conn)
		}()
	}

	// wait all the task to end when the error exists
	s.wait.Wait()
}

func (s *TcpServer) Stop() {
	s.isClosed = true
	s.closeCh <- struct{}{}
}
