package database

import (
	"github.com/xzwsloser/Go-redis/lib/utils"
	"log"
	"testing"
	"time"
)

func TestPExpireAt(t *testing.T) {
	db := NewDatabase(0)
	c1 := [][]byte{
		[]byte("k1"),
		[]byte("v1"),
	}

	c2 := [][]byte{
		[]byte("k1"),
	}

	c3 := utils.ExpireCmd("k1", time.Now().Add(time.Second*5))[1:]

	execSet(db, c1)
	execPExpireAt(db, c3)

	for i := 0; i < 10; i++ {
		reply := execGet(db, c2)
		log.Println("======", i, "======")
		log.Print(string(reply.ToByte()))
		log.Println("===============")
		time.Sleep(time.Second)
	}
}
