package aof

import (
	"github.com/xzwsloser/Go-redis/interface/database"
)

const (
	STRING_SET_COMMAND = "SET"
)

func EntityToCmd(key string, data *database.DataEntity) [][]byte {
	switch data.Data.(type) {
	case []byte:
		return newStringCmd(key, data.Data.([]byte))
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
