package database

import (
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"log"
	"sync"
	"testing"
	"time"
)

func TestMIncr(t *testing.T) {
	db := NewDatabase(0)
	key := []byte("key")
	value := []byte("0")
	execSet(db, [][]byte{key, value})

	var wg sync.WaitGroup
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		go func() {
			defer wg.Done()
			execIncr(db, [][]byte{key})
		}()
	}

	wg.Wait()
	reply := execGet(db, [][]byte{key})
	log.Print(string(reply.ToByte()))
}

func TestMSet(t *testing.T) {
	db := NewDatabase(0)
	mset(db)
}

func mset(db *Database) {
	cmdLine := [][]byte{
		[]byte("k1"),
		[]byte("v1"),
		[]byte("k2"),
		[]byte("v2"),
		[]byte("k3"),
		[]byte("v3"),
		[]byte("k4"),
		[]byte("v4"),
	}

	reply := execMSet(db, cmdLine)
	log.Println("===========")
	log.Print(string(reply.ToByte()))
	log.Println("===========")
}

func TestMGet(t *testing.T) {
	db := NewDatabase(0)
	mset(db)
	cmdLine := [][]byte{
		[]byte("k1"),
		[]byte("k2"),
		[]byte("k3"),
		[]byte("k4"),
	}

	reply := execMGet(db, cmdLine)
	log.Println("===========")
	log.Print(string(reply.ToByte()))
	log.Println("===========")
}

func TestSetEx(t *testing.T) {
	db := NewDatabase(0)
	var command = [][]byte{
		[]byte("k1"),
		[]byte("v1"),
		[]byte("5"),
	}

	var commandGet = [][]byte{
		[]byte("k1"),
	}

	reply := execSetEx(db, command)
	log.Print(string(reply.ToByte()))
	for i := 0; i < 8; i++ {
		reply = execGet(db, commandGet)
		if protocol.IsErrReply(reply) {
			log.Println("key not exists")
		} else {
			log.Println("find the key")
		}
		time.Sleep(time.Second)
	}
}
