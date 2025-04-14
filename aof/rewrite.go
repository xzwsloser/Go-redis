package aof

import (
	"errors"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/lib/utils"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"io"
	"os"
	"strconv"
)

const (
	AOF_REWRITE_TEMP_NAME = "temp.aof"
)

type RewriteCtx struct {
	tmpFile  *os.File
	fileSize int64
	dbIndex  int
}

func (persister *Persister) newRewriteHandler() *Persister {
	h := &Persister{}
	h.aofFileName = persister.aofFileName
	h.db = persister.tmpDBMaker()
	return h
}

// Rewrite do the rewrite operation
func (persister *Persister) Rewrite() (err error) {
	ctx, err := persister.PreRewrite()
	if err != nil {
		return
	}
	err = persister.DoRewrite(ctx)
	if err != nil {
		return
	}
	err = persister.FinishRewrite(ctx)
	return
}

// DoRewrite call the real rewrite function
func (persister *Persister) DoRewrite(ctx *RewriteCtx) (err error) {
	return persister.generatorAof(ctx)
}

// Rewrite
func (persister *Persister) PreRewrite() (ctx *RewriteCtx, err error) {
	persister.pauseLock.Lock()
	defer persister.pauseLock.Unlock()

	// record the current size of the aof file
	stat, err := os.Stat(persister.aofFileName)
	if err != nil {
		logger.Error("failed to open aof file! err: %s", err.Error())
		return
	}
	fileSize := stat.Size()
	tmpFile, err := os.OpenFile(AOF_REWRITE_TEMP_NAME,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0644)
	if err != nil {
		logger.Error("failed to open temp file pointer err: %s", err.Error())
		return
	}

	ctx = &RewriteCtx{
		fileSize: fileSize,
		tmpFile:  tmpFile,
	}
	return
}

func (persister *Persister) FinishRewrite(ctx *RewriteCtx) (err error) {
	tmpFile := ctx.tmpFile
	if tmpFile == nil {
		logger.Error("tmp file fd is not open")
		return errors.New("tmp file fd is not open")
	}

	file, err := os.OpenFile(persister.aofFileName,
		os.O_RDONLY,
		0644)
	if err != nil {
		logger.Error("open aof file failed err: %s", err.Error())
		return err
	}

	_, err = file.Seek(ctx.fileSize, 0)
	if err != nil {
		logger.Error("update the new position of the pointer failed! err: %s", err.Error())
		return err
	}

	selectCmd := protocol.
		NewMultiReply(utils.CmdLine1("SELECT",
			strconv.Itoa(ctx.dbIndex))).ToByte()

	_, err = tmpFile.Write(selectCmd)
	if err != nil {
		logger.Error("write into temp aof filed failed err: %s", err.Error())
		return err
	}

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		logger.Error("write extra content to temp aof file failed!")
		return
	}

	err = os.Rename(tmpFile.Name(), persister.aofFileName)
	if err != nil {
		logger.Error("rename the file: %s to file: %s failed Err: %s",
			tmpFile.Name(), persister.aofFileName, err.Error())
		return err
	}

	file, err = os.OpenFile(persister.aofFileName,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644)
	if err != nil {
		panic(err)
	}

	persister.aofWriter = file
	selectCmd = protocol.
		NewMultiReply(utils.CmdLine1("SELECT",
			strconv.Itoa(persister.curDBIndex))).ToByte()
	_, err = persister.aofWriter.Write(selectCmd)
	if err != nil {
		panic(err)
	}
	return nil
}
