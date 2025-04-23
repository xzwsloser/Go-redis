package database

import (
	"log"
	"testing"
)

func TestListPush(t *testing.T) {
	command := [][]byte{
		[]byte("list"),
		[]byte("1"),
		[]byte("2"),
		[]byte("3"),
		[]byte("4"),
	}
	db := NewDatabase(0)
	reply := execLPush(db, command)
	log.Print(reply)
}

func TestLRem(t *testing.T) {
	command := [][]byte{
		[]byte("list"),
		[]byte("hello1"),
		[]byte("hello2"),
		[]byte("hello3"),
		[]byte("hello4"),
		[]byte("hello4444"),
		[]byte("hello5555"),
		[]byte("hello6666"),
		[]byte("hello4"),
		[]byte("hello5"),
		[]byte("hello6"),
		[]byte("hello7"),
		[]byte("hello8"),
		[]byte("hello9"),
		[]byte("hello10"),
	}
	db := NewDatabase(0)
	execRPush(db, command)
	command = [][]byte{
		[]byte("list"),
		[]byte("0"),
		[]byte("hello4"),
	}
	reply := execLRem(db, command)
	log.Print(string(reply.ToByte()))
}

func TestLRange(t *testing.T) {
	command := [][]byte{
		[]byte("list"),
		[]byte("hello1"),
		[]byte("hello2"),
		[]byte("hello3"),
		[]byte("hello4"),
		[]byte("hello4444"),
		[]byte("hello5555"),
		[]byte("hello6666"),
		[]byte("hello4"),
		[]byte("hello5"),
		[]byte("hello6"),
		[]byte("hello7"),
		[]byte("hello8"),
		[]byte("hello9"),
		[]byte("hello10"),
	}
	db := NewDatabase(0)
	execRPush(db, command)
	command = [][]byte{
		[]byte("list"),
		[]byte("2"),
		[]byte("6"),
	}
	reply := execLRange(db, command)
	log.Print(string(reply.ToByte()))
}
