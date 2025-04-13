package server

import (
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"log"
	"net"
	"sync"
	"testing"
)

// test string command
func TestSet(t *testing.T) {
	server := NewTcpServer()
	go server.Run()

	var command1 = "*3\r\n$3\r\nSET\r\n$3\r\nKEY\r\n$5\r\nVALUE\r\n"
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Error(err)
	}

	_, _ = conn.Write([]byte(command1))
	var result []byte = make([]byte, 100)
	n, _ := conn.Read(result)
	log.Println("=====received=====")
	log.Print(string(result[:n]))
	log.Println("============")
}

func TestGet(t *testing.T) {
	server := NewTcpServer()
	go server.Run()

	var command1 = "*3\r\n$3\r\nSET\r\n$3\r\nKEY\r\n$5\r\nVALUE\r\n"
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Error(err)
	}

	_, _ = conn.Write([]byte(command1))
	var result []byte = make([]byte, 100)
	n, _ := conn.Read(result)
	log.Println("=====received=====")
	log.Print(string(result[:n]))
	log.Println("============")

	var command2 = [][]byte{
		[]byte("GET"),
		[]byte("KEY"),
	}

	getCommand := protocol.NewMultiReply(command2).ToByte()
	log.Println(string(getCommand))
	_, _ = conn.Write(getCommand)
	n, _ = conn.Read(result)
	log.Println("=====getResult=====")
	log.Print(string(result[:n]))
	log.Println("============")
}

func TestIncr(t *testing.T) {
	server := NewTcpServer()
	go server.Run()

	var command1 = [][]byte{
		[]byte("Set"),
		[]byte("number"),
		[]byte("22"),
	}

	cmd1 := protocol.NewMultiReply(command1).ToByte()
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Error(err)
	}

	_, _ = conn.Write(cmd1)
	var result []byte = make([]byte, 100)
	n, _ := conn.Read(result)
	log.Println("=====received=====")
	log.Print(string(result[:n]))
	log.Println("============")

	var command2 = [][]byte{
		[]byte("Incr"),
		[]byte("number"),
	}

	getCommand := protocol.NewMultiReply(command2).ToByte()
	log.Println(string(getCommand))
	_, _ = conn.Write(getCommand)
	n, _ = conn.Read(result)
	log.Println("=====getResult=====")
	log.Print(string(result[:n]))
	log.Println("============")

	var command3 = [][]byte{
		[]byte("Get"),
		[]byte("number"),
	}
	getCommand1 := protocol.NewMultiReply(command3).ToByte()
	log.Print(string(getCommand1))
	_, _ = conn.Write(getCommand1)
	n, _ = conn.Read(result)
	log.Println("=====getResult=====")
	log.Print(string(result[:n]))
	log.Println("============")
}

func setNumber(conn net.Conn) {
	var command1 = [][]byte{
		[]byte("Set"),
		[]byte("number"),
		[]byte("0"),
	}

	cmd1 := protocol.NewMultiReply(command1).ToByte()
	_, _ = conn.Write(cmd1)
	var result []byte = make([]byte, 100)
	n, _ := conn.Read(result)
	log.Println("=====received=====")
	log.Print(string(result[:n]))
	log.Println("============")
}

func TestMutliIncr(t *testing.T) {
	server := NewTcpServer()
	go server.Run()
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		t.Error(err)
	}

	var command1 = [][]byte{
		[]byte("Set"),
		[]byte("number"),
		[]byte("0"),
	}

	cmd1 := protocol.NewMultiReply(command1).ToByte()
	_, _ = conn.Write(cmd1)
	var result []byte = make([]byte, 100)
	n, _ := conn.Read(result)
	log.Println("=====received=====")
	log.Print(string(result[:n]))
	log.Println("============")

	var incrCommand = [][]byte{
		[]byte("Incr"),
		[]byte("number"),
	}

	commandToSend := protocol.NewMultiReply(incrCommand).ToByte()

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(k int) {
			defer func() {
				wg.Done()
				if v := recover(); v != nil {
					log.Print("error")
				}
			}()

			c, err := net.Dial("tcp", ":8080")
			if err != nil {
				return
			}

			_, _ = c.Write(commandToSend)

		}(i)
	}

	wg.Wait()

	var getCommand = [][]byte{
		[]byte("Get"),
		[]byte("number"),
	}

	commandGetToSend := protocol.NewMultiReply(getCommand).ToByte()
	_, _ = conn.Write(commandGetToSend)
	NewResult := make([]byte, 100)
	n, _ = conn.Read(NewResult)
	log.Println("=========r========")
	log.Print(string(NewResult[:n]))
	log.Println("==================")
}

func mset(conn net.Conn) {
	commands := [][]byte{
		[]byte("MSET"),
		[]byte("k1"),
		[]byte("v1"),
		[]byte("k2"),
		[]byte("v2"),
		[]byte("k3"),
		[]byte("v3"),
		[]byte("k4"),
		[]byte("v4"),
	}

	commandToSend := protocol.NewMultiReply(commands).ToByte()
	_, _ = conn.Write(commandToSend)
	result := make([]byte, 100)
	n, _ := conn.Read(result)
	log.Println("=====received=====")
	log.Print(string(result[:n]))
	log.Println("=================")
}

func TestMSet(t *testing.T) {
	server := NewTcpServer()
	go server.Run()

}
