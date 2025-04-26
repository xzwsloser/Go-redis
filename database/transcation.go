package database

import (
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"strings"
)

func init() {
	RegisterCommand("GETVERSION", execGetVersion, readFirstKey, nil, 2)
}

func Watch(db *Database, conn redis.Conn, args [][]byte) redis.Reply {
	if conn != nil && conn.InitMulti() {
		return protocol.NewErrReply("ERR WATCH inside MULTI is not allowed")
	}
	watching := conn.GetWatching()
	for _, arg := range args {
		key := string(arg)
		versionCode := db.GetVersion(key)
		watching[key] = versionCode
	}
	return protocol.NewOkReply()
}

func execGetVersion(db *Database, args [][]byte) redis.Reply {
	key := string(args[0])
	versionCode := db.GetVersion(key)
	return protocol.NewIntReply(int64(versionCode))
}

func isWatchingChanged(db *Database, watching map[string]uint32) bool {
	for key, version := range watching {
		currentVersion := db.GetVersion(key)
		if version != currentVersion {
			return true
		}
	}
	return false
}

func StartMulti(conn redis.Conn) redis.Reply {
	if conn.InitMulti() {
		return protocol.NewErrReply("not allow multi-transcation")
	}
	conn.SetMulti(true)
	return protocol.NewOkReply()
}

func (db *Database) GetUndoLogs(cmdLine [][]byte) []CmdLine {
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := commandTable[cmdName]
	if !ok {
		return nil
	}
	if cmd.undo == nil {
		return nil
	}
	return cmd.undo(db, cmdLine[1:])
}

func EnqueueCmd(conn redis.Conn, args [][]byte) redis.Reply {
	cmdName := strings.ToLower(string(args[0]))
	cmd, ok := commandTable[cmdName]
	if !ok {
		err := protocol.NewErrReply("not exists command")
		conn.AddTxErrors(err)
		return err
	}
	if cmd.prepare == nil {
		err := protocol.NewErrReply("not exists prepare function")
		conn.AddTxErrors(err)
		return err
	}

	if validCommand(args) != nil {
		err := protocol.NewErrReply("in valid args of the command")
		conn.AddTxErrors(err)
		return err
	}
	conn.EnqueueCmd(args)
	return protocol.NewQueuedReply()
}

func ExecMulti(db *Database, conn redis.Conn) redis.Reply {
	if !conn.InitMulti() {
		return protocol.NewErrReply("not in the state of the transcation")
	}
	defer conn.SetMulti(false)
	if len(conn.GetTxErrors()) > 0 {
		return protocol.NewErrReply("ERROR occurred in the process of the enqueue")
	}
	return execMulti(db, conn)
}

func execMulti(db *Database, conn redis.Conn) redis.Reply {
	// 1. prepare write keys and read keys
	wks := make([]string, 0)
	rks := make([]string, 0)
	commands := conn.GetCmdLineInQueue()
	for _, command := range commands {
		cmdName := strings.ToLower(string(command[0]))
		cmd, ok := commandTable[cmdName]
		if !ok || cmd.prepare == nil {
			continue
		}
		wk, rk := cmd.prepare(command[1:])
		wks = append(wks, wk...)
		rks = append(rks, rk...)
	}
	// 1.2 prepare the watching keys
	watchKeys := make([]string, 0)
	watching := conn.GetWatching()
	for watch, _ := range watching {
		watchKeys = append(watchKeys, watch)
	}
	rks = append(rks, watchKeys...)
	// 1.3 lock the keys
	db.RWLocks(wks, rks)
	defer db.RWUnlocks(wks, rks)
	if isWatchingChanged(db, watching) {
		return protocol.NewEmptyReply()
	}

	// 2. exec the commands and get the undo logs
	result := make([]redis.Reply, 0, len(commands))
	abort := false
	undoCmdLines := make([][]CmdLine, 0, len(commands))
	for _, command := range commands {
		undoCmdLines = append(undoCmdLines, db.GetUndoLogs(command))
		reply := db.execWithLock(conn, command)
		if protocol.IsErrReply(reply) {
			abort = true
			undoCmdLines = undoCmdLines[:len(undoCmdLines)-1]
			break
		}
		result = append(result, reply)
	}
	if !abort {
		db.AddVersion(wks...)
		return protocol.NewMultiRawReply(result)
	}
	// 3. undo the commands
	for i := len(undoCmdLines) - 1; i >= 0; i-- {
		undoCmd := undoCmdLines[i]
		if len(undoCmd) == 0 {
			continue
		}
		for _, command := range undoCmd {
			_ = db.execWithLock(conn, command)
		}
	}
	return protocol.NewErrReply("ERROR occurred in the exec process of the trancation")
}

func DiscardMulti(db *Database, conn redis.Conn) redis.Reply {
	if !conn.InitMulti() {
		return protocol.NewErrReply("ERROR there is no transcation to discard")
	}
	conn.ClearCmdQueue()
	conn.SetMulti(false)
	return protocol.NewOkReply()
}
