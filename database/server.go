package database

import (
	"errors"
	"github.com/xzwsloser/Go-redis/aof"
	"github.com/xzwsloser/Go-redis/config"
	"github.com/xzwsloser/Go-redis/interface/database"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/pub"
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
	dbSet     []*atomic.Value
	persister *aof.Persister
	hub       *pub.Hub
}

func init() {
	RegisterCommand("ping", execPing, nil, nil, 1)
}

func NewRedisServer() *RedisServer {
	dbNumber := config.GetDBConfig().Number
	if dbNumber <= 0 {
		dbNumber = 16
	}

	dbSet := make([]*atomic.Value, dbNumber)
	for i := 0; i < dbNumber; i++ {
		dbSet[i] = &atomic.Value{}
		dbSet[i].Store(NewDatabase(i))
	}

	server := &RedisServer{
		dbSet: dbSet,
	}

	persister := aof.NewPersister()
	persister.SetTmpDBMaker(func() database.DBEngine {
		return NewPureServer()
	})
	if persister != nil {
		persister.BindRedisServer(server)
		if persister.Load {
			persister.LoadAof()
		}
		server.bindPersister(persister)
	}
	server.hub = pub.NewHub()
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
	} else if cmdName == "bgwriteaof" {
		r.execBgReWrite()
		return protocol.NewOkReply()
	} else if cmdName == "subscribe" {
		return r.hub.Subscribe(conn, cmdLine[1:])
	} else if cmdName == "unsubscribe" {
		return r.hub.Unsubscribe(conn, cmdLine[1:])
	} else if cmdName == "publish" {
		return r.hub.Publish(cmdLine[1:])
	}

	return r.execNormalCommand(conn, cmdLine)
}

func (r *RedisServer) execNormalCommand(conn redis.Conn, cmdLine [][]byte) redis.Reply {
	cmdName := strings.ToLower(string(cmdLine[0]))
	err := validCommand(cmdLine)
	if err != nil {
		return protocol.NewErrReply(err.Error())
	}

	selectDB, err := r.selectDB(conn.GetDBIndex())
	if err != nil {
		return protocol.NewErrReply(err.Error())
	}

	cmd, ok := commandTable[cmdName]
	if !ok {
		return protocol.NewErrReply(COMMAND_NOT_FIND)
	}

	prepare := cmd.prepare
	if prepare != nil {
		wks, rks := prepare(cmdLine[1:])
		selectDB.RWLocks(wks, rks)
		defer selectDB.RWUnlocks(wks, rks)
	}
	reply := cmd.exector(selectDB, cmdLine[1:])
	if reply == nil {
		return protocol.NewErrReply(EMPTY_REPLY)
	}
	return reply
}

func (r *RedisServer) execWithLock(conn redis.Conn, cmdLine [][]byte) redis.Reply {
	cmdName := string(cmdLine[0])
	err := validCommand(cmdLine)
	if err != nil {
		return protocol.NewErrReply(err.Error())
	}

	selectDB, err := r.selectDB(conn.GetDBIndex())
	if err != nil {
		return protocol.NewErrReply(err.Error())
	}

	cmd, ok := commandTable[cmdName]
	if !ok {
		return protocol.NewErrReply(COMMAND_NOT_FIND)
	}

	reply := cmd.exector(selectDB, cmdLine[1:])
	if reply == nil {
		return protocol.NewErrReply(EMPTY_REPLY)
	}
	return reply
}

func (r *RedisServer) Close() {
	if r.persister != nil {
		r.persister.Close()
	}
}

func (r *RedisServer) AfterClientClose(conn redis.Conn) {
	r.hub.UnSubscribeAll(conn)
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

func (s *RedisServer) mustSelectDB(index int) *Database {
	var db *Database
	db, err := s.selectDB(index)
	if err != nil {
		db, err = s.selectDB(0)
		if err != nil {
			db = NewDatabase(0)
		}
	}
	return db
}

// ForEach Scan all the k-v in database
func (s *RedisServer) ForEach(dbIndex int, consumer func(key string, value *database.DataEntity) bool) {
	s.mustSelectDB(dbIndex).ForEach(consumer)
}

func (s *RedisServer) ReadAOF() {
	err := s.persister.Rewrite()
	if err != nil {
		logger.Error("rewrite failed!")
	}
}

func (s *RedisServer) execBgReWrite() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("err: %v", err)
			}
		}()
		s.ReadAOF()
	}()
}

func NewPureServer() *RedisServer {
	dbNum := config.GetDBConfig().Number
	if dbNum <= 0 {
		dbNum = 16
	}
	server := &RedisServer{}
	server.dbSet = make([]*atomic.Value, dbNum)
	for i := 0; i < dbNum; i++ {
		server.dbSet[i] = &atomic.Value{}
		server.dbSet[i].Store(NewDatabase(i))
	}
	return server
}
