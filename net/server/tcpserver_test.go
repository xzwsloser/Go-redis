package server

import (
	"bufio"
	"log"
	"net"
	"testing"
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
