package aof

import (
	"github.com/xzwsloser/Go-redis/datastruct/list"
	"github.com/xzwsloser/Go-redis/datastruct/sortedset"
	"github.com/xzwsloser/Go-redis/interface/database"
	"strconv"
)

const (
	STRING_SET_COMMAND  = "SET"
	LIST_PUSH_COMMAND   = "RPUSH"
	ZSET_INSERT_COMMAND = "ZADD"
)

func EntityToCmd(key string, data *database.DataEntity) [][]byte {
	switch data.Data.(type) {
	case []byte:
		return newStringCmd(key, data.Data.([]byte))
	case *list.LinkedList:
		return newListCmd(key, data.Data.(*list.LinkedList))
	case *sortedset.SortedSet:
		return newSortedSet(key, data.Data.(*sortedset.SortedSet))
	default:
		return [][]byte{}
	}
}

func newStringCmd(key string, value []byte) [][]byte {
	result := make([][]byte, 3)
	result[0] = []byte(STRING_SET_COMMAND)
	result[1] = []byte(key)
	result[2] = value
	return result
}

func newListCmd(key string, value *list.LinkedList) [][]byte {
	result := make([][]byte, 2+value.Len())
	result[0] = []byte(LIST_PUSH_COMMAND)
	result[1] = []byte(key)
	i := 0
	value.ForEach(func(key any) bool {
		result[i+2] = key.([]byte)
		i++
		return true
	})
	return result
}

func newSortedSet(key string, value *sortedset.SortedSet) [][]byte {
	result := make([][]byte, 2+value.Len()*2)
	result[0] = []byte(ZSET_INSERT_COMMAND)
	result[1] = []byte(key)
	i := 2
	value.ForEach(func(score float64, key string) bool {
		result[i] = []byte(strconv.FormatFloat(score, 'f', 2, 10))
		result[i+1] = []byte(key)
		i += 2
		return true
	})
	return result
}
