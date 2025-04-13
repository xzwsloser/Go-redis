package database

import (
	"errors"
	"github.com/xzwsloser/Go-redis/config"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	ARGS_OF_COMMAND_ERR = "args of command err"
	COMMAND_NOT_FIND    = "command not find"
	DB_NOT_FIND         = "database not find"
	DB_INDEX_ERR        = "database index is not valid"
	EMPTY_REPLY         = "the reply is empty"
)

// RedisServer is the inner server to exec command  like the httpServer
type RedisServer struct {
	dbSet []*atomic.Value
}

func init() {
	RegisterCommand("ping", execPing, 1)
}

func NewRedisServer() *RedisServer {
	dbNumber := config.GetDBConfig().Number
	if dbNumber <= 0 {
		dbNumber = 16
	}

	dbSet := make([]*atomic.Value, dbNumber)
	for i := 0; i < dbNumber; i++ {
		dbSet[i] = &atomic.Value{}
		dbSet[i].Store(NewDatabase())
	}

	server := &RedisServer{
		dbSet: dbSet,
	}
	return server
}

func execPing(db *Database, cmdLine [][]byte) redis.Reply {
	return protocol.NewStatusReply("PONG")
}

func (r *RedisServer) Exec(conn redis.Conn, cmdLine [][]byte) redis.Reply {
	cmdName := strings.ToLower(string(cmdLine[0]))
	if cmdName == "select" {
		if len(cmdLine) != 2 {
			return protocol.NewErrReply(ARGS_OF_COMMAND_ERR)
		}

		indexStr := string(cmdLine[1])
		index, err := strconv.ParseInt(indexStr, 10, 32)
		if err != nil || index < 0 || index > int64(len(r.dbSet)) {
			return protocol.NewErrReply(DB_INDEX_ERR)
		}
		conn.SelectDB(int(index))
	}

	err := validCommand(cmdLine)
	if err != nil {
		return protocol.NewErrReply(err.Error())
	}

	seleteDB, err := r.selectDB(conn.GetDBIndex())
	if err != nil {
		return protocol.NewErrReply(err.Error())
	}

	cmd, ok := commandTable[cmdName]
	if !ok {
		return protocol.NewErrReply(COMMAND_NOT_FIND)
	}

	reply := cmd.exector(seleteDB, cmdLine[1:])
	if reply == nil {
		return protocol.NewErrReply(EMPTY_REPLY)
	}
	return reply
}

func (r *RedisServer) Close() {

}

func (r *RedisServer) AfterClientClose(conn redis.Conn) {

}

func validCommand(commandLine [][]byte) error {
	commandName := strings.ToLower(string(commandLine[0]))
	commandInfo, exists := commandTable[commandName]
	if !exists {
		return errors.New(COMMAND_NOT_FIND)
	}

	if commandInfo.arity < 0 {
		if len(commandLine) < -commandInfo.arity {
			return errors.New(ARGS_OF_COMMAND_ERR)
		}
	} else {
		if len(commandLine) != commandInfo.arity {
			return errors.New(ARGS_OF_COMMAND_ERR)
		}
	}

	return nil
}

func (s *RedisServer) selectDB(index int) (*Database, error) {
	if index < 0 || index >= len(s.dbSet) {
		return nil, errors.New(DB_NOT_FIND)
	}

	db, ok := s.dbSet[index].Load().(*Database)
	if !ok {
		return nil, errors.New(DB_NOT_FIND)
	}
	return db, nil
}
