package database

import (
	"github.com/xzwsloser/Go-redis/aof"
	"github.com/xzwsloser/Go-redis/lib/utils"
)

// writeFirstKey: get the key of the write command  example: set k1 v1
func writeFirstKey(args [][]byte) ([]string, []string) {
	key := string(args[0])
	return []string{key}, nil
}

// readFirstKey: get the key of the read command example: get k1
func readFirstKey(args [][]byte) ([]string, []string) {
	key := string(args[0])
	return nil, []string{key}
}

// writeKeys: get many write keys , example: DEL k1 , k2 , k3 ...
func writeKeys(args [][]byte) ([]string, []string) {
	wks := make([]string, len(args))
	for i, arg := range args {
		wks[i] = string(arg)
	}
	return wks, nil
}

// readKeys: get many read keys , example: MGet k1 , k2 , k3 ...
func readKeys(args [][]byte) ([]string, []string) {
	rks := make([]string, len(args))
	for i, arg := range args {
		rks[i] = string(arg)
	}
	return nil, rks
}

func rollbackFirstKey(db *Database, args [][]byte) []CmdLine {
	key := string(args[0])
	return rollbackGivenKeys(db, key)
}

func rollbackGivenKeys(db *Database, keys ...string) []CmdLine {
	undoCmdLine := make([][][]byte, 0)
	for _, key := range keys {
		entity, exists := db.GetEntityWithLock(key)
		if !exists {
			undoCmdLine = append(undoCmdLine,
				utils.CmdLine1("DEL", key))
		} else {
			undoCmdLine = append(undoCmdLine,
				utils.CmdLine1("DEL", key),
				aof.EntityToCmd(key, entity),
				db.TTLCmd(key))
		}
	}
	return undoCmdLine
}
