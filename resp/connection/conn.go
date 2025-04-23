package connection

import (
	"github.com/xzwsloser/Go-redis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

var (
	MaxTimeOut = time.Second * 10
)

type Connection struct {
	mu            *sync.Mutex
	conn          net.Conn
	sendDataWait  wait.Wait
	selectDBIndex int
	channels      map[string]bool
}

func (c *Connection) Subscribe(channel string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.channels[channel]
	if ok {
		return false
	}
	c.channels[channel] = true
	return true
}

func (c *Connection) UnSubscribe(channel string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.channels[channel]; !ok {
		return false
	}
	delete(c.channels, channel)
	return true
}

func (c *Connection) GetChannel() []string {
	res := make([]string, len(c.channels))
	c.mu.Lock()
	defer c.mu.Unlock()
	i := 0
	for channel, _ := range c.channels {
		res[i] = channel
		i++
	}
	return res
}

func (c *Connection) SubsCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.channels)
}

func (c *Connection) GetDBIndex() int {
	return c.selectDBIndex
}

func (c *Connection) SelectDB(dbIndex int) {
	c.selectDBIndex = dbIndex
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:     conn,
		mu:       &sync.Mutex{},
		channels: make(map[string]bool),
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
