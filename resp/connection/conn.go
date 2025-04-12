package connection

import (
	"github.com/xzwsloser/Go-redis/lib/sync/wait"
	"net"
	"time"
)

var (
	MaxTimeOut = time.Second * 10
)

type Connection struct {
	conn          net.Conn
	sendDataWait  wait.Wait
	selectDBIndex int
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

func (c *Connection) Write(msg []byte) (int, error) {
	c.sendDataWait.Add(1)
	defer func() {
		c.sendDataWait.Done()
	}()
	return c.conn.Write(msg)
}

func (c *Connection) Close() error {
	c.sendDataWait.WaitWithTimeout(MaxTimeOut)
	return c.conn.Close()
}

func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}
