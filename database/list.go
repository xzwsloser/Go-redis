package database

import (
	"github.com/xzwsloser/Go-redis/datastruct/list"
	"github.com/xzwsloser/Go-redis/interface/database"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"strconv"
)

func init() {
	RegisterCommand("LINDEX", execLIndex, readFirstKey, nil, 3)
	RegisterCommand("LLEN", execLen, readFirstKey, nil, 2)
	RegisterCommand("LPOP", execLPop, writeFirstKey, nil, 2)
	RegisterCommand("LPUSH", execLPush, writeFirstKey, nil, -3)
	RegisterCommand("RPOP", execRPop, writeFirstKey, nil, 2)
	RegisterCommand("RPUSH", execRPush, writeFirstKey, nil, -3)
	RegisterCommand("LREM", execLRem, writeFirstKey, nil, 4)
	RegisterCommand("LRANGE", execLRange, readFirstKey, nil, 4)
}

func (db *Database) getOrInitLinkedList(key string) *list.LinkedList {
	entity, exists := db.GetEntityWithLock(key)
	if !exists {
		ll := list.NewLinkedList()
		db.PutEntityWithLock(key, &database.DataEntity{
			Data: ll,
		})
		return ll
	}
	ll, ok := entity.Data.(*list.LinkedList)
	if !ok {
		ll := list.NewLinkedList()
		db.PutEntityWithLock(key, &database.DataEntity{
			Data: ll,
		})
		return ll
	}
	return ll
}

// LIndex key 0
func execLIndex(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	valueStr := string(cmdLine[1])
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return protocol.NewErrReply("in valid argement of LINDEX")
	}
	ll := db.getOrInitLinkedList(key)
	res := ll.Get(value)
	if res == nil {
		return protocol.NewErrReply("in valid range of the index in LINDEX")
	}
	return protocol.NewBulkReply([]byte(res.(string)))
}

// LLEN key
func execLen(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	ll := db.getOrInitLinkedList(key)
	l := ll.Len()
	return protocol.NewIntReply(int64(l))
}

// LPOP list
func execLPop(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	ll := db.getOrInitLinkedList(key)
	value := ll.RemoveHead()
	if value == nil {
		return protocol.NewBulkReply([]byte("nil"))
	}
	return protocol.NewIntReply(int64(value.(int)))
}

// LPUSH list v1 v2 v3 ...
func execLPush(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	ll := db.getOrInitLinkedList(key)
	for i := 0; i < len(cmdLine)-1; i++ {
		ll.InsertHead(string(cmdLine[i+1]))
	}
	return protocol.NewIntReply(int64(ll.Len()))
}

// RPop list
func execRPop(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	ll := db.getOrInitLinkedList(key)
	value := ll.RemoveTail()
	if value == nil {
		return protocol.NewBulkReply([]byte("nil"))
	}
	return protocol.NewBulkReply([]byte(value.(string)))
}

// RPUSH list v1 v2 ...
func execRPush(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	ll := db.getOrInitLinkedList(key)
	for i := 0; i < len(cmdLine)-1; i++ {
		ll.InsertTail(string(cmdLine[i+1]))
	}
	return protocol.NewIntReply(int64(ll.Len()))
}

// LRem list count value
func execLRem(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	countStr := string(cmdLine[1])
	value := string(cmdLine[2])
	ll := db.getOrInitLinkedList(key)
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return protocol.NewErrReply("invalid args of LREM")
	}

	var res int
	if count > 0 {
		res = ll.RemoveByValue(value, false)
	} else if count < 0 {
		res = ll.RemoveByValue(value, true)
	} else {
		temp := ll.RemoveByValue(value, false)
		res += temp
		for temp != 0 {
			temp = ll.RemoveByValue(value, false)
			res += temp
		}
	}
	return protocol.NewIntReply(int64(res))
}

// LRange list start stop
func execLRange(db *Database, cmdLine [][]byte) redis.Reply {
	key := string(cmdLine[0])
	start, err := strconv.Atoi(string(cmdLine[1]))
	stop, err := strconv.Atoi(string(cmdLine[2]))
	if err != nil {
		return protocol.NewErrReply("invalid number for LRANGE")
	}
	ll := db.getOrInitLinkedList(key)
	values := ll.FindRangeValue(start, stop)
	if len(values) == 0 {
		return protocol.NewBulkReply([]byte("empty list"))
	}
	res := make([][]byte, len(values))
	for i, v := range values {
		res[i] = []byte(strconv.Itoa(i+start) + ") " + v.(string))
	}
	return protocol.NewMultiReply(res)
}
