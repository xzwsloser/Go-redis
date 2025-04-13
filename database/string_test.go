package database

import (
	"log"
	"sync"
	"testing"
)

func TestMIncr(t *testing.T) {
	db := NewDatabase()
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
	db := NewDatabase()
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
	db := NewDatabase()
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
