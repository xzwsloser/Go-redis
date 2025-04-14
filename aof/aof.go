package aof

import (
	"context"
	"errors"
	"github.com/xzwsloser/Go-redis/config"
	"github.com/xzwsloser/Go-redis/interface/database"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/lib/utils"
	"github.com/xzwsloser/Go-redis/resp/connection"
	"github.com/xzwsloser/Go-redis/resp/parse"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CmdLine = [][]byte

const (
	AOF_CHAN_SIZE = 1 << 20
	BUF_SIZE      = 1 << 10 // 1 KB
)

const (
	AOF_FSYNC_ALWAYS = "always"
	AOF_FSYNC_SECOND = "everysec"
	AOF_FSYNC_NO     = "no"
)

type Persister struct {
	db         database.DBEngine
	tmpDBMaker func() database.DBEngine
	// ctx and cancel is as a signal to end the listening goroutinue
	ctx    context.Context
	cancel context.CancelFunc
	// aofChan is the chan for goroutinue to fetch payLoad and write into buffer
	aofChan     chan *payLoad
	aofFinished chan struct{}
	aofWriter   *os.File
	// aofFsync is the way to write aof
	aofFsync string
	// aofFileName is the name of aof file
	aofFileName string
	buffer      []CmdLine
	pauseLock   sync.Mutex
	curDBIndex  int
	AppendOnly  bool
	Load        bool
}

type payLoad struct {
	dbIndex int
	cmdLine CmdLine
	w       sync.WaitGroup
}

func NewPersister() *Persister {
	aofChan := make(chan *payLoad, AOF_CHAN_SIZE)
	aofFinished := make(chan struct{})
	buf := make([]CmdLine, BUF_SIZE)
	aofFileName := config.GetAofConfig().AppendFileName
	aofFileSync := config.GetAofConfig().AppendFileSync
	aofFileWriter, err := os.OpenFile(aofFileName,
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644)
	if err != nil {
		logger.Error("failed to open append file: %v", err.Error())
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())

	persister := &Persister{
		ctx:         ctx,
		cancel:      cancel,
		aofChan:     aofChan,
		aofFinished: aofFinished,
		aofWriter:   aofFileWriter,
		aofFsync:    aofFileSync,
		aofFileName: aofFileName,
		buffer:      buf,
	}

	if config.GetAofConfig().AppendOnly == "on" {
		persister.AppendOnly = true
	}

	if config.GetAofConfig().Load == "on" {
		persister.Load = true
	}

	go persister.listenCmd()

	if persister.aofFsync == AOF_FSYNC_SECOND {
		go persister.fsyncEverySecond()
	}

	return persister
}

func (persister *Persister) SetTmpDBMaker(maker func() database.DBEngine) {
	persister.tmpDBMaker = maker
}

func (persister *Persister) BindRedisServer(db database.DBEngine) {
	persister.db = db
}

func (persister *Persister) listenCmd() {
	for msg := range persister.aofChan {
		persister.writeAof(msg)
	}
	persister.aofFinished <- struct{}{}
}

func (persister *Persister) SaveCmdLine(dbIndex int, cmdLine CmdLine) {
	if cmdLine == nil {
		return
	}

	if persister.aofFsync == AOF_FSYNC_ALWAYS {
		p := &payLoad{
			dbIndex: dbIndex,
			cmdLine: cmdLine,
		}
		persister.writeAof(p)
		return
	}

	persister.aofChan <- &payLoad{
		dbIndex: dbIndex,
		cmdLine: cmdLine,
	}
}

func (persister *Persister) writeAof(p *payLoad) {
	// keep the capacity
	persister.buffer = persister.buffer[:0]
	persister.pauseLock.Lock()
	defer persister.pauseLock.Unlock()
	if p.dbIndex != persister.curDBIndex {
		selectDBCmd := utils.CmdLine1("SELECT", strconv.Itoa(p.dbIndex))
		persister.buffer = append(persister.buffer, selectDBCmd)
		data := protocol.NewMultiReply(selectDBCmd).ToByte()
		_, err := persister.aofWriter.Write(data)
		if err != nil {
			logger.Error("write into aof file buffer err: %v", err.Error())
			return
		}
		persister.curDBIndex = p.dbIndex
	}

	persister.buffer = append(persister.buffer, p.cmdLine)
	data := protocol.NewMultiReply(p.cmdLine).ToByte()
	_, err := persister.aofWriter.Write(data)
	if err != nil {
		logger.Error("write into aof file buffer err: %v", err.Error())
		return
	}

	if persister.aofFsync == AOF_FSYNC_ALWAYS {
		err := persister.aofWriter.Sync()
		if err != nil {
			logger.Error("sync into aof file err: %v", err.Error())
		}
	}
}

func (persister *Persister) Close() {
	if persister.aofWriter != nil {
		close(persister.aofChan)
		<-persister.aofFinished
		err := persister.aofWriter.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}
	persister.cancel()
}

func (persister *Persister) Fsync() {
	persister.pauseLock.Lock()
	defer persister.pauseLock.Unlock()
	err := persister.aofWriter.Sync()
	if err != nil {
		logger.Error("sync into aof file err: %v", err.Error())
		return
	}
}

func (persister *Persister) fsyncEverySecond() {
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			persister.Fsync()
		case <-persister.ctx.Done():
			return
		}
	}
}

func (persister *Persister) LoadAof() {
	file, err := os.OpenFile(persister.aofFileName,
		os.O_RDONLY,
		0644)
	if err != nil {
		logger.Error("failed to open file err: %v", err)
		return
	}

	var reader io.Reader
	reader = file

	fake := connection.NewFakeConnection()
	ch := parse.ParseStream(reader)
	for payLoad := range ch {
		if payLoad.Error != nil {
			if payLoad.Error == io.EOF {
				logger.Warn("the connection is closed")
				return
			}
			logger.Error("load aof file err: %v", payLoad.Error.Error())
			continue
		}

		if payLoad.Data == nil {
			logger.Warn("empty payLoad Err: %v", payLoad.Error.Error())
			continue
		}

		reply, ok := payLoad.Data.(*protocol.MulitBulkReply)
		if !ok {
			logger.Warn("not a executable command")
			continue
		}

		newReply := persister.db.Exec(fake, reply.Args)
		if protocol.IsErrReply(newReply) {
			logger.Error("executor err")
			continue
		}

		if strings.ToLower(string(reply.Args[0])) == "select" {
			dbIndex, err := strconv.ParseInt(string(reply.Args[1]), 10, 64)
			if err != nil {
				logger.Warn("valid db index")
				continue
			}
			persister.curDBIndex = int(dbIndex)
		}
	}
}

// generatorAof generate the aof file into the temp file
func (persister *Persister) generatorAof(ctx *RewriteCtx) error {
	persister.pauseLock.Lock()
	defer persister.pauseLock.Unlock()
	handler := persister.newRewriteHandler()
	tmpFile := ctx.tmpFile
	if tmpFile == nil {
		logger.Warn("temp file fd is not open")
		return errors.New("temp file fd is not open")
	}
	handler.LoadAof()
	// rewrite the aof file
	selectCmd := utils.CmdLine1("SELECT", strconv.Itoa(ctx.dbIndex))
	selectReply := protocol.NewMultiReply(selectCmd)
	_, err := tmpFile.Write(selectReply.ToByte())
	if err != nil {
		logger.Error("write select command into database failed")
		return err
	}
	handler.db.ForEach(ctx.dbIndex, func(key string, value *database.DataEntity) bool {
		cmdLine := EntityToCmd(key, value)
		cmd := protocol.NewMultiReply(cmdLine).ToByte()
		_, err = tmpFile.Write(cmd)
		if err != nil {
			logger.Error("failed to rewrite message into temp aof file err: %s", err.Error())
			return false
		}
		return true
	})
	return nil
}
