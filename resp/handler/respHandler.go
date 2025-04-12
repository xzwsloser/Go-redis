package handler

import (
	"bufio"
	"context"
	"github.com/xzwsloser/Go-redis/interface/database"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/lib/sync/atomic"
	"github.com/xzwsloser/Go-redis/resp/connection"
	"github.com/xzwsloser/Go-redis/resp/parse"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"io"
	"net"
	"strings"
	"sync"
)

const (
	ERR_USE_CLOSED_NETWORK string = "use of closed network connection"
)

// RespHandler contains the callback function to deal with resp protocol
type RespHandler struct {
	activeConn sync.Map
	db         database.DB
	isClosing  atomic.Boolean
}

func NewRespHandler() *RespHandler {
	r := &RespHandler{}
	r.isClosing.Set(false)
	return r
}

func (r *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if r.isClosing.Get() {
		_ = conn.Close()
		return
	}

	client := connection.NewConnection(conn)
	r.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	payLoads := parse.ParseStream(reader)
	for payLoad := range payLoads {
		if payLoad.Error != nil {
			if payLoad.Error == io.EOF ||
				payLoad.Error == io.ErrUnexpectedEOF ||
				strings.Contains(payLoad.Error.Error(), ERR_USE_CLOSED_NETWORK) {
				// close the connection
				r.closeSingleClient(client)
				logger.Error("respHandler occur io Err: %v",
					payLoad.Error.Error())
				return
			}

			// other error
			logger.Error("respHandler occur Err: %v",
				payLoad.Error.Error())
			errReply := protocol.NewErrReply(payLoad.Error.Error())
			_, _ = client.Write(errReply.ToByte())
			continue
		}

		// invoke the db.Exec function
		if payLoad.Data != nil {
			logger.Error("respHandler received empty request")
			errReply := protocol.NewErrReply("empty request")
			_, _ = client.Write(errReply.ToByte())
			continue
		}

		request, ok := payLoad.Data.(*protocol.MulitBulkReply)
		if !ok {
			logger.Error("respHandler received a non-MultiBulkReply")
			errReply := protocol.NewErrReply("need MulitBulkReply")
			_, _ = client.Write(errReply.ToByte())
			continue
		}

		reply := r.db.Exec(client, request.Args)

		if reply == nil {
			reply = protocol.NewUnknownReply()
		}

		_, _ = conn.Write(reply.ToByte())
	}
}

func (r *RespHandler) Close() error {
	r.isClosing.Set(true)
	r.activeConn.Range(func(k, v any) bool {
		_ = k.(redis.Conn).Close()
		return true
	})

	r.db.Close()
	return nil
}

func (r *RespHandler) closeSingleClient(client redis.Conn) {
	_ = client.Close()
	r.db.AfterClientClose(client)
	r.activeConn.Delete(client)
}
