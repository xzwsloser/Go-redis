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
			execDecr(db, [][]byte{key})
		}()
	}

	wg.Wait()
	reply := execGet(db, [][]byte{key})
	log.Print(string(reply.ToByte()))
}
