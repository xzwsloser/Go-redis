package redis

// Conn is the connection between client and redis server
type Conn interface {
	Write([]byte) (int, error)
	Close() error
	RemoteAddr() string
	GetDBIndex() int
	SelectDB(int)
	Subscribe(channel string) bool
	UnSubscribe(channel string) bool
	GetChannel() []string
	SubsCount() int
	GetWatching() map[string]uint32
	InitMulti() bool
	SetMulti(bool)
	EnqueueCmd([][]byte)
	GetCmdLineInQueue() [][][]byte
	ClearCmdQueue()
	AddTxErrors(err error)
	GetTxErrors() []error
}
