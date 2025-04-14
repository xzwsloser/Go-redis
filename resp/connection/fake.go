package connection

const (
	BUFFER_SIZE = 1 << 10
)

// FakeConnection fake connection to load aof file and a buffer (no use...)
type FakeConnection struct {
	Connection
	buf    []byte
	offset int
}

func NewFakeConnection() *FakeConnection {
	fake := &FakeConnection{}
	fake.buf = make([]byte, 0, BUFFER_SIZE)
	fake.offset = 0
	return fake
}

func (fake *FakeConnection) Write(data []byte) (n int, err error) {
	if len(data)+fake.offset > cap(fake.buf) {
		new_buf := make([]byte, 0, cap(fake.buf)*2)
		new_buf = append(new_buf, fake.buf...)
		fake.buf = new_buf
	}

	fake.buf = append(fake.buf, data...)
	fake.offset += len(data)
	return len(data), nil
}

func (fake *FakeConnection) Read(p []byte) (n int, err error) {
	n = copy(p, fake.buf[:fake.offset])
	return n, nil
}

func (fake *FakeConnection) Clear() {
	fake.buf = nil
	fake.offset = 0
}

func (fake *FakeConnection) Close() error {
	fake.Clear()
	return nil
}
