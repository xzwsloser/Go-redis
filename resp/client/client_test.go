package client

import (
	"log"
	"strconv"
	"testing"
)

func TestRedisClient(t *testing.T) {
	client, err := NewRedisClient("localhost:8080")
	if err != nil {
		t.Error(err)
	}
	client.Start()
	var command = [][]byte{
		[]byte("SET"),
		[]byte("key1"),
		[]byte("value1"),
	}
	reply := client.Send(command)
	log.Print(string(reply.ToByte()))
}

func TestMultiCommand(t *testing.T) {
	client, err := NewRedisClient("localhost:8080")
	if err != nil {
		t.Error(err)
	}
	client.Start()
	for i := 0; i < 100; i++ {
		command := [][]byte{
			[]byte("SET"),
			[]byte("key" + strconv.Itoa(i)),
			[]byte("value" + strconv.Itoa(i)),
		}
		reply := client.Send(command)
		log.Print(reply)
	}
}
