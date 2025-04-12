package redis

// Conn is the connection between client and redis server
type Conn interface {
	Write([]byte) (int, error)
	Close() error
	RemoteAddr() string
}
