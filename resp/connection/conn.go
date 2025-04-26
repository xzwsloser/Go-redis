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

const (
	flagMulti uint64 = 1 << iota
)

type Connection struct {
	mu            *sync.Mutex
	conn          net.Conn
	sendDataWait  wait.Wait
	selectDBIndex int
	// the channel the client has subscribed
	channels map[string]bool
	// key -> versionCode
	watching map[string]uint32
	// flags: the flags of the current state of the client
	flags uint64
	// queue: the queue of the command send in the state of the transcation
	queue  [][][]byte
	txErrs []error
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
		watching: make(map[string]uint32),
		flags:    0,
		queue:    make([][][]byte, 0),
		txErrs:   make([]error, 0),
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
	c.channels = nil
	c.watching = nil
	c.mu = nil
	return c.conn.Close()
}

func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *Connection) GetWatching() map[string]uint32 {
	if c.watching == nil {
		c.watching = make(map[string]uint32)
	}
	return c.watching
}

// InitMulit: judge is there already has transcation
func (c *Connection) InitMulti() bool {
	return c.flags&flagMulti > 0
}

func (c *Connection) SetMulti(state bool) {
	if !state {
		c.queue = nil
		c.watching = nil
		c.flags &= ^flagMulti
		return
	}
	c.flags |= flagMulti
}

func (c *Connection) EnqueueCmd(cmdLine [][]byte) {
	c.queue = append(c.queue, cmdLine)
}

func (c *Connection) GetCmdLineInQueue() [][][]byte {
	return c.queue
}

func (c *Connection) ClearCmdQueue() {
	c.queue = nil
}

func (c *Connection) AddTxErrors(err error) {
	c.txErrs = append(c.txErrs, err)
}

func (c *Connection) GetTxErrors() []error {
	return c.txErrs
}
