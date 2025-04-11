package protocol

import (
	"bytes"
	"github.com/xzwsloser/Go-redis/interface/redis"
	"strconv"
)

/**
this file record many type of resp protocol example
1. Status Reply:
	+OK/r/n
2. Err Reply:
	-Err message\r\n
3. Int Reply:
	:12\r\n
4. string reply (BulkReply)
	$4\r\nPING\r\n
5. multi-string reply
	*3\r\n$3\r\nSET\r\n$3\r\nKEY\r\n$5\r\nVALUE\r\n
*/

const (
	CRLF = "\r\n"
)

// StatusReply bring some infos
type StatusReply struct {
	status string
}

func NewStatusReply(status string) *StatusReply {
	return &StatusReply{
		status: status,
	}
}

func (s *StatusReply) ToByte() []byte {
	return []byte("+" + s.status + CRLF)
}

var okReply = NewStatusReply("OK")

func NewOkReply() *StatusReply {
	return okReply
}

func IsOkReply(reply redis.Reply) bool {
	buf := reply.ToByte()
	return string(buf[1:len(buf)-2]) == "OK"
}

// ErrReply bing the info of error
type ErrReply struct {
	errorMsg string
}

func NewErrReply(errMsg string) *ErrReply {
	return &ErrReply{
		errorMsg: errMsg,
	}
}

func (e *ErrReply) ToByte() []byte {
	return []byte("-" + e.errorMsg + CRLF)
}

func IsErrReply(reply redis.Reply) bool {
	return reply.ToByte()[0] == '-'
}

// IntReply bing int
type IntReply struct {
	value int64
}

func NewIntReply(value int64) *IntReply {
	return &IntReply{
		value: value,
	}
}

func (i *IntReply) ToByte() []byte {
	return []byte(":" + strconv.FormatInt(i.value, 10) + CRLF)
}

func IsIntReply(reply redis.Reply) bool {
	return reply.ToByte()[0] == ':'
}

// BulkReply wrap string
type BulkReply struct {
	content []byte
}

func NewBulkReply(content []byte) *BulkReply {
	return &BulkReply{
		content: content,
	}
}

func (b *BulkReply) ToByte() []byte {
	lStr := strconv.Itoa(int(len(b.content)))
	return []byte("$" + lStr + CRLF + string(b.content) + CRLF)
}

// MulitBulkReply is the many line reply
type MulitBulkReply struct {
	args [][]byte
}

func NewMultiReply(args [][]byte) *MulitBulkReply {
	return &MulitBulkReply{
		args: args,
	}
}

func (m *MulitBulkReply) ToByte() []byte {
	argLen := len(m.args)
	var buf bytes.Buffer
	buf.WriteString("*")
	buf.WriteString(strconv.Itoa(argLen))
	buf.WriteString(CRLF)
	for _, arg := range m.args {
		buf.WriteString("$")
		buf.WriteString(strconv.Itoa(len(arg)))
		buf.WriteString(CRLF)
		buf.Write(arg)
		buf.WriteString(CRLF)
	}
	return buf.Bytes()
}

type EmptyReply struct{}

func NewEmptyReply() *EmptyReply {
	return &EmptyReply{}
}

func (e *EmptyReply) ToByte() []byte {
	return []byte("*0\r\n")
}
