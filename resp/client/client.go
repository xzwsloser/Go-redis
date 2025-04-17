package client

import (
	"errors"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/lib/sync/wait"
	"github.com/xzwsloser/Go-redis/resp/parse"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"net"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

/**
@Author: loser
@Description: 一个Redis客户端连接
*/

const (
	closing uint32 = 0
	running uint32 = 1
)

const (
	headBeat uint = 0
	normal   uint = 1
)

const (
	MaxWaitTimeOut   = 3 * time.Second
	MaxHeartBeatTime = 5 * time.Second
	MaxChanSize      = 1 << 8
)

type RedisClient struct {
	conn       net.Conn // TCP 连接
	status     uint32
	reqToSend  chan *Request // 等待发送的消息队列
	reqToReply chan *Request // 等待回复的消息队列
	waiting    *wait.Wait    // 使用 waitGroup 优雅关闭连接
	ticker     *time.Ticker
	addr       string
}

type Request struct {
	requestType uint
	args        [][]byte
	reply       redis.Reply
	waiting     *wait.Wait
	err         error
}

func NewRedisClient(addr string) (*RedisClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	client := &RedisClient{
		status:     closing,
		reqToSend:  make(chan *Request, MaxChanSize),
		reqToReply: make(chan *Request, MaxChanSize),
		waiting:    &wait.Wait{},
		addr:       addr,
		conn:       conn,
	}
	return client, nil
}

func (client *RedisClient) Start() {
	client.ticker = time.NewTicker(MaxHeartBeatTime)
	atomic.StoreUint32(&client.status, running)
	go client.handleWrite()
	go client.handleRead()
	go client.handleHeartBeat()
}

func (client *RedisClient) Close() {
	atomic.StoreUint32(&client.status, closing)
	client.ticker.Stop()
	close(client.reqToSend)
	client.waiting.Wait()
	close(client.reqToReply)
	_ = client.conn.Close()
}

// @brief: 向连接中写入数据
func (client *RedisClient) handleWrite() {
	for req := range client.reqToSend {
		client.doRequest(req)
	}
}

// @brief: 处理 handlWrite 读出来的信息
func (client *RedisClient) doRequest(req *Request) {
	if req == nil || len(req.args) == 0 {
		return
	}
	var err error
	reqContent := protocol.NewMultiReply(req.args).ToByte()
	for i := 0; i < 3; i++ {
		_, err = client.conn.Write(reqContent)
		if err == nil ||
			(!strings.Contains(err.Error(), "timeout") &&
				!strings.Contains(err.Error(), "deadline exceeded")) {
			break
		}
		time.Sleep(time.Second)
		continue
	}

	if err == nil {
		client.reqToReply <- req
	} else {
		req.err = err
		req.waiting.Done()
	}
}

// @brief: 从网络连接中读取信息,并且解析信息发送给客户端
func (client *RedisClient) handleRead() {
	ch := parse.ParseStream(client.conn)
	for payLoad := range ch {
		if payLoad.Error != nil {
			if atomic.LoadUint32(&client.status) == closing {
				return
			}
			client.reconnect()
			return
		}
		client.finishRequest(payLoad.Data)
	}
}

// @brief: 完成请求,从等待回复队列中获取到请求对象并且封装返回
func (client *RedisClient) finishRequest(reply redis.Reply) {
	defer func() {
		if v := recover(); v != nil {
			debug.PrintStack()
			logger.Error("Err: ", v)
		}
	}()
	req := <-client.reqToReply
	if req == nil {
		return
	}
	req.reply = reply
	if req.waiting != nil {
		req.waiting.Done()
	}
}

// @brief: 心跳检测执行函数
func (client *RedisClient) handleHeartBeat() {
	for range client.ticker.C {
		client.doHeartBeat()
	}
}

// @brief: 实际执行心跳检测的函数
func (client *RedisClient) doHeartBeat() {
	req := &Request{
		requestType: headBeat,
		args:        [][]byte{[]byte("PING")},
		waiting:     &wait.Wait{},
		err:         nil,
	}

	client.waiting.Add(1)
	defer client.waiting.Done()
	req.waiting.Add(1)
	client.reqToSend <- req
	req.waiting.WaitWithTimeout(MaxWaitTimeOut)
}

func (client *RedisClient) Send(args [][]byte) redis.Reply {
	if atomic.LoadUint32(&client.status) == closing {
		return protocol.NewErrReply("the client is closed")
	}
	req := &Request{
		requestType: normal,
		args:        args,
		waiting:     &wait.Wait{},
		err:         nil,
	}
	client.waiting.Add(1)
	defer client.waiting.Done()
	req.waiting.Add(1)
	client.reqToSend <- req
	timeout := req.waiting.WaitWithTimeout(MaxWaitTimeOut)
	if timeout {
		return protocol.NewErrReply("redis server timeout")
	}

	if req.err != nil {
		return protocol.NewErrReply("req err: " + req.err.Error())
	}
	return req.reply
}

func (client *RedisClient) reconnect() {
	logger.Error("begin to reconnect ...")
	_ = client.conn.Close()
	var conn net.Conn
	var err error
	for i := 0; i < 3; i++ {
		conn, err = net.Dial("tcp", client.addr)
		if err != nil {
			time.Sleep(time.Second)
			continue
		} else {
			break
		}
	}

	if conn == nil {
		client.Close()
		return
	}

	close(client.reqToReply)
	for req := range client.reqToReply {
		req.err = errors.New("reconnect to redis server")
		req.waiting.Done()
	}
	client.conn = conn
	client.reqToReply = make(chan *Request, MaxChanSize)
	go client.handleRead()
}
