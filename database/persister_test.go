package database

import (
	"github.com/xzwsloser/Go-redis/aof"
	"github.com/xzwsloser/Go-redis/resp/connection"
	"log"
	"testing"
)

func TestPersister(t *testing.T) {
	persister := aof.NewPersister()
	db := NewDatabase()
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
