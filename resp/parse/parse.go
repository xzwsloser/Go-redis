package parse

import (
	"bufio"
	"errors"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"github.com/xzwsloser/Go-redis/lib/logger"
	"github.com/xzwsloser/Go-redis/resp/protocol"
	"io"
	"strconv"
)

/**
1. +OK\r\n
2. -Err message\r\n
3. :100\r\n
4. $4\r\nPING\r\n
5. *3\r\n$3\r\nSET\r\n$3\r\nKEY\r\n$5\r\nVALUE\r\n
*/

const (
	Protocol_Err_Message = "Protocol Err: "
)

type PayLoad struct {
	Data  redis.Reply
	Error error
}

type readState struct {
	isParsingLine bool
	expectedArgs  int
	args          [][]byte
	bulkLen       int
	msgType       byte
}

func (s *readState) finished() bool {
	return s.expectedArgs > 0 && s.expectedArgs == len(s.args)
}

func NewProtocolError(msg string) error {
	return errors.New(Protocol_Err_Message + msg)
}

func ParseStream(reader io.Reader) <-chan *PayLoad {
	ch := make(chan *PayLoad, 5)
	go func() {
		parse0(reader, ch)
	}()
	return ch
}

// parse0 the core of state machine
func parse0(reader io.Reader, ch chan<- *PayLoad) {
	// in this goroutinue  do not let the error throwed to the main goroutinue
	defer func() {
		if err := recover(); err != nil {
			logger.Error("parse0 err: %v", err)
		}
	}()

	bufioReader := bufio.NewReader(reader)
	var state readState
	state.isParsingLine = false
	for {
		msg, ioErr, err := readLine(bufioReader, &state)
		if err != nil {
			// if the connection is closed , close the channel
			if ioErr {
				logger.Error("parse0 IO Err: %s", err.Error())
				close(ch)
				return
			}

			logger.Error("parse0 readLine Err: %v", err.Error())

			ch <- &PayLoad{
				Error: err,
			}

			state = readState{}
			continue
		}

		// not begin to parse a new message
		if !state.isParsingLine {
			if msg[0] == '*' {
				err := parseMultiBulkMsg(msg, &state)
				if err != nil {
					logger.Error("parse multi header Err: %v", err.Error())
					ch <- &PayLoad{
						Error: err,
					}
					state = readState{}
					continue
				}

				if state.expectedArgs == 0 {
					ch <- &PayLoad{
						Data: protocol.NewEmptyReply(),
					}

					state = readState{}
					continue
				}

			} else if msg[0] == '$' {
				err := parseBulkMsg(msg, &state)
				if err != nil {
					ch <- &PayLoad{
						Error: err,
					}

					state = readState{}
					continue
				}
			} else {
				reply, err := parseSingleLine(msg)
				if err != nil {
					ch <- &PayLoad{
						Data:  reply,
						Error: err,
					}
					continue
				}

				ch <- &PayLoad{
					Data: reply,
				}
			}
		} else {
			// is parsing the content after *n\r\n
			err := readBody(msg, &state)
			if err != nil {
				ch <- &PayLoad{
					Error: err,
				}

				state = readState{}
				continue
			}

			if state.finished() {
				if state.msgType == '*' {
					reply := protocol.NewMultiReply(state.args)
					ch <- &PayLoad{
						Data: reply,
					}

					state = readState{}
				} else if state.msgType == '$' {
					reply := protocol.NewBulkReply(state.args[0])
					ch <- &PayLoad{
						Data: reply,
					}

					state = readState{}
				}
			}

		}
	}
}

// readLine parse the context of a single line by the bulklen of the state(juding the header of the content)
func readLine(reader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var (
		msg []byte
		err error
	)

	if state.bulkLen == 0 {
		msg, err = reader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}

		if len(msg) < 1 || msg[len(msg)-1] != '\n' {
			return nil, false, NewProtocolError(string(msg))
		}
	} else {
		msg = make([]byte, state.bulkLen+2)
		_, err = io.ReadFull(reader, msg)
		if err != nil {
			return nil, true, err
		}

		if len(msg) < 2 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, NewProtocolError(string(msg))
		}

		state.bulkLen = 0
	}

	return msg, false, nil
}

// parseSingleLine parse single line e.g +OK\r\n :123\r\n ...
func parseSingleLine(msg []byte) (redis.Reply, error) {
	var res redis.Reply
	if msg[0] == '+' {
		status := msg[1 : len(msg)-2]
		res = protocol.NewStatusReply(string(status))
	} else if msg[0] == '-' {
		errMsg := msg[1 : len(msg)-2]
		res = protocol.NewErrReply(string(errMsg))
	} else {
		value := msg[1 : len(msg)-2]
		val, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return nil, NewProtocolError(string(msg))
		}
		res = protocol.NewIntReply(val)
	}
	return res, nil
}

// parseBulkMsg method parse the request e.g $4\r\n
func parseBulkMsg(msg []byte, state *readState) error {
	msgLenStr := string(msg[1 : len(msg)-2])
	msgLen, err := strconv.Atoi(msgLenStr)
	if err != nil {
		return NewProtocolError(string(msg))
	}

	state.bulkLen = msgLen
	state.isParsingLine = true
	state.expectedArgs = 1
	state.args = make([][]byte, 0, 1)
	state.msgType = '$'

	return nil
}

// parseMultiBulkMsg parse the Multi Header e.g *3\r\n
func parseMultiBulkMsg(msg []byte, state *readState) error {
	lineStr := string(msg[1 : len(msg)-2])
	argsLen, err := strconv.Atoi(lineStr)
	if err != nil {
		return NewProtocolError(string(msg))
	}

	state.expectedArgs = argsLen
	state.isParsingLine = true
	state.args = make([][]byte, 0, argsLen)
	state.msgType = '*'
	return nil
}

// readBody read the next info of the header    e.g PING\r\n  or $4\r\n
func readBody(msg []byte, state *readState) error {
	var contentStr string
	if msg[0] == '$' {
		contentStr = string(msg[1 : len(msg)-2])
		contentLen, err := strconv.Atoi(contentStr)
		if err != nil {
			return NewProtocolError(string(msg))
		}

		state.bulkLen = contentLen
	} else {
		contentStr = string(msg[0 : len(msg)-2])
		state.args = append(state.args, []byte(contentStr))
	}
	return nil
}
