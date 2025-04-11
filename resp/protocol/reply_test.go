package protocol

import (
	"github.com/xzwsloser/Go-redis/interface/redis"
	"log"
	"testing"
)

func TestGenerateReply(t *testing.T) {
	statusReply := NewStatusReply("Hello World")
	disPlay(statusReply, "status")
	okReply := NewOkReply()
	disPlay(okReply, "ok")
	errReply := NewErrReply("Err message")
	disPlay(errReply, "error")
	intReply := NewIntReply(100)
	disPlay(intReply, "int")
	strReply := NewBulkReply([]byte("PING"))
	disPlay(strReply, "string")
	args := [][]byte{
		[]byte("SET"),
		[]byte("KEY"),
		[]byte("VALUE"),
	}

	multiStrReply := NewMultiReply(args)
	disPlay(multiStrReply, "muti-string")
}

func disPlay(reply redis.Reply, description string) {
	log.Println("=====" + description + "=====")
	log.Print(string(reply.ToByte()))
	log.Println("===================")
}
