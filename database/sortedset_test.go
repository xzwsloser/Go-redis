package database

import (
	"log"
	"testing"
)

func TestZRangeByRank(t *testing.T) {
	db := NewDatabase(0)
	cmdLine := [][]byte{
		[]byte("key1"),
		[]byte("1.0"),
		[]byte("a"),
		[]byte("2.0"),
		[]byte("b"),
		[]byte("3.0"),
		[]byte("c"),
		[]byte("4.0"),
		[]byte("d"),
	}
	reply := execZadd(db, cmdLine)
	log.Print(string(reply.ToByte()))

	cmdLine = [][]byte{
		[]byte("key1"),
		[]byte("-1"),
		[]byte("5"),
	}
	reply = execZrange(db, cmdLine)
	log.Print(string(reply.ToByte()))
}

func TestZRemByRankRange(t *testing.T) {
	db := NewDatabase(0)
	cmdLine := [][]byte{
		[]byte("key1"),
		[]byte("1.0"),
		[]byte("a"),
		[]byte("2.0"),
		[]byte("b"),
		[]byte("3.0"),
		[]byte("c"),
		[]byte("4.0"),
		[]byte("d"),
	}
	reply := execZadd(db, cmdLine)
	log.Print(string(reply.ToByte()))

	cmdLine = [][]byte{
		[]byte("key1"),
		[]byte("-1"),
		[]byte("5"),
	}
	reply = execZRemRangeByRank(db, cmdLine)
	log.Print(string(reply.ToByte()))
}
