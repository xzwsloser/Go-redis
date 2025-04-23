package pub

import (
	"github.com/xzwsloser/Go-redis/datastruct/list"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/utils"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"strconv"
)

var (
	_subscribe         = "subscribe"
	_unsubscribe       = "unsubscribe"
	messageBytes       = []byte("message")
	unSubscribeNothing = []byte("*3\r\n$11\r\nunsubscribe\r\n$-1\n:0\r\n")
)

// makeMsg: make the message send to the clients subscribed the channel
func makeMsg(t string, channel string, code int64) []byte {
	return []byte("*3\r\n$" + strconv.FormatInt(int64(len(t)), 10) + protocol.CRLF + t + protocol.CRLF +
		"$" + strconv.FormatInt(int64(len(channel)), 10) + protocol.CRLF + channel + protocol.CRLF +
		":" + strconv.FormatInt(code, 10) + protocol.CRLF)
}

func (hub *Hub) subscribe0(channel string, c redis.Conn) {
	ok := c.Subscribe(channel)
	// already subscribe the channel
	if !ok {
		return
	}
	raw, exists := hub.subs.Get(channel)
	if exists {
		ll := raw.(*list.LinkedList)
		ll.InsertTail(c)
		return
	}
	ll := list.NewLinkedList()
	ll.InsertTail(c)
	hub.subs.Put(channel, ll)
}

func (hub *Hub) unsubscribe0(channel string, c redis.Conn) {
	ok := c.UnSubscribe(channel)
	// not subscribe
	if !ok {
		return
	}
	raw, exists := hub.subs.Get(channel)
	if !exists {
		return
	}
	linkedList := raw.(*list.LinkedList)
	linkedList.RemoveByCond(func(idx int, value any) bool {
		return utils.Equals(value, c)
	})
	if linkedList.Len() == 0 {
		hub.subs.Remove(channel)
	}
}

func (hub *Hub) Unsubscribe(c redis.Conn, args [][]byte) redis.Reply {
	channels := make([]string, len(args))
	for i, arg := range args {
		channels[i] = string(arg)
	}
	hub.lockers.Locks(channels)
	defer hub.lockers.Unlocks(channels)
	for _, channel := range channels {
		hub.unsubscribe0(channel, c)
		_, _ = c.Write(makeMsg(_unsubscribe, channel, int64(c.SubsCount())))
	}
	return protocol.NewNoReply()
}

func (hub *Hub) Subscribe(c redis.Conn, args [][]byte) redis.Reply {
	channels := make([]string, len(args))
	for i, arg := range args {
		channels[i] = string(arg)
	}
	hub.lockers.Locks(channels)
	defer hub.lockers.Unlocks(channels)
	for _, channel := range channels {
		hub.subscribe0(channel, c)
		_, _ = c.Write(makeMsg(_subscribe, channel, int64(c.SubsCount())))
	}
	return protocol.NewNoReply()
}

func (hub *Hub) UnSubscribeAll(c redis.Conn) redis.Reply {
	channels := c.GetChannel()
	if len(channels) == 0 {
		_, _ = c.Write(unSubscribeNothing)
		return protocol.NewNoReply()
	}
	hub.lockers.Locks(channels)
	defer hub.lockers.Unlocks(channels)
	for _, channel := range channels {
		hub.unsubscribe0(channel, c)
		_, _ = c.Write(makeMsg(_unsubscribe, channel, int64(c.SubsCount())))
	}
	return protocol.NewNoReply()
}

func (hub *Hub) Publish(args [][]byte) redis.Reply {
	if len(args) != 2 {
		return protocol.NewErrReply("args of Publish err")
	}
	channel := string(args[0])
	message := args[1]
	hub.lockers.Locks([]string{channel})
	defer hub.lockers.Unlocks([]string{channel})
	raw, ok := hub.subs.Get(channel)
	if !ok {
		return protocol.NewErrReply("no such a channel")
	}
	linkedlist := raw.(*list.LinkedList)
	linkedlist.ForEach(func(value any) bool {
		c := value.(redis.Conn)
		replyArgs := make([][]byte, 3)
		replyArgs[0] = messageBytes
		replyArgs[1] = []byte(channel)
		replyArgs[2] = message
		_, _ = c.Write(protocol.NewMultiReply(replyArgs).ToByte())
		return true
	})
	return protocol.NewIntReply(int64(linkedlist.Len()))
}
