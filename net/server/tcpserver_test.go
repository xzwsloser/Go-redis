package server

import (
	"bufio"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func TestTcpServer(t *testing.T) {
	server := NewTcpServer()
	go server.Run()
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Fatal(err.Error())
	}

	_, _ = conn.Write([]byte("Hello\n"))
	var buf [100]byte
	reader := bufio.NewReader(conn)
	n, _ := reader.Read(buf[:])
	if string(buf[:n]) != "Hello\n" {
		t.Error("echo failed")
	}
	log.Print(string(buf[:n]))
	server.Stop()
}

func TestPubSub(t *testing.T) {
	server := NewTcpServer()
	go server.Run()
	c1, err := net.Dial("tcp", ":8080")
	c2, err := net.Dial("tcp", ":8080")
	c3, err := net.Dial("tcp", ":8080")
	c4, err := net.Dial("tcp", ":8080")
	subscribers := []net.Conn{c2, c3, c4}
	if err != nil {
		t.Error(err)
	}

	w := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	w.Add(4)
	for i := 0; i < 3; i++ {
		go func(idx int) {
			defer w.Done()
			// 订阅连接
			command1 := [][]byte{
				[]byte("subscribe"),
				[]byte("channel1"),
			}

			_, _ = subscribers[idx].Write(protocol.NewMultiReply(command1).ToByte())

			if idx == 1 {
				command3 := [][]byte{
					[]byte("unsubscribe"),
					[]byte("channel1"),
				}
				_, _ = subscribers[idx].Write(protocol.NewMultiReply(command3).ToByte())
			}

			b := make([]byte, 100)
			for {
				n, err := subscribers[idx].Read(b)
				if err == io.EOF {
					break
				}
				mu.Lock()
				log.Println("======subscriber=====", idx)
				log.Print(string(b[:n]))
				log.Println("=====================")
				mu.Unlock()
			}
		}(i)
	}

	time.Sleep(time.Second)
	go func() {
		defer w.Done()
		command2 := [][]byte{
			[]byte("publish"),
			[]byte("channel1"),
			[]byte("hello world"),
		}
		_, _ = c1.Write(protocol.NewMultiReply(command2).ToByte())
		b := make([]byte, 100)
		n, err := c1.Read(b)
		if err == io.EOF {
			return
		}
		mu.Lock()
		log.Println("========publisher========")
		log.Print(string(b[:n]))
		log.Println("=========================")
		mu.Unlock()
	}()

	w.Wait()
}

func TestPreFunc(t *testing.T) {
	server := NewTcpServer()
	go server.Run()
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Error(err)
	}
	command1 := [][]byte{
		[]byte("SET"),
		[]byte("k100"),
		[]byte("v100"),
	}
	_, _ = conn.Write(protocol.NewMultiReply(command1).ToByte())
	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		t.Error(err)
	}
	log.Println("=====result of the first command=====")
	log.Print(string(buf[:n]))
	log.Println("=====end of the first command=====")
	command2 := [][]byte{
		[]byte("GET"),
		[]byte("k100"),
	}
	_, _ = conn.Write(protocol.NewMultiReply(command2).ToByte())
	buf = make([]byte, 100)
	n, err = conn.Read(buf)
	if err != nil {
		t.Error(err)
	}
	log.Println("=====result of the second command=====")
	log.Print(string(buf[:n]))
	log.Println("=====end of the second command=====")
}
