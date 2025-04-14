package database

import (
	"github.com/xzwsloser/Go-redis/aof"
	"github.com/xzwsloser/Go-redis/resp/connection"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestPersister(t *testing.T) {
	persister := aof.NewPersister()
	db := NewDatabase(0)
	db.addAof = func(cmdLine [][]byte) {
		if persister.AppendOnly {
			persister.SaveCmdLine(db.index, cmdLine)
		}
	}

	var commandSet = [][]byte{
		[]byte("k1"),
		[]byte("v1"),
	}
	execSet(db, commandSet)

	db.index = 2
	var commandSet1 = [][]byte{
		[]byte("k2"),
		[]byte("0"),
	}
	execSet(db, commandSet1)

	var commandIncr = [][]byte{
		[]byte("k2"),
	}
	execIncr(db, commandIncr)
}

func TestLoadAof(t *testing.T) {
	server := NewRedisServer()
	var commandInfo = [][]byte{
		[]byte("GET"),
		[]byte("k1"),
	}
	fake := connection.NewFakeConnection()
	reply := server.Exec(fake, commandInfo)
	log.Print(string(reply.ToByte()))
}

func TestGeneratorData(t *testing.T) {
	persister := aof.NewPersister()
	commands := make([][]byte, 101)
	commands[0] = []byte("HSET")
	j := 0
	for i := 0; i < 100; i += 2 {
		commands[i+1] = []byte("k" + strconv.Itoa(j))
		commands[j+2] = []byte("v" + strconv.Itoa(j))
		j++
	}
	db := NewDatabase(0)
	db.addAof = func(cmdLine [][]byte) {
		persister.SaveCmdLine(db.index, cmdLine)
	}
	execMSet(db, commands[1:])
}

func TestRewrite(t *testing.T) {
	server := NewRedisServer()
	ti := time.Now()
	server.ReadAOF()
	log.Println("time: ", time.Since(ti).Nanoseconds())
}

func TestBgWriteAof(t *testing.T) {
	server := NewRedisServer()
	var commands = [][]byte{
		[]byte("BGWRITEAOF"),
	}
	server.Exec(connection.NewFakeConnection(), commands)
	time.Sleep(time.Second)
}
