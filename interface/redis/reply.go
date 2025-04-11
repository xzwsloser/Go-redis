package redis

type Reply interface {
	ToByte() []byte
}
