package parse

import (
	"bytes"
	"log"
	"testing"
)

func TestParseMsgSingleLine(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	ch := ParseStream(bytes.NewReader(buf.Bytes()))
	for v := range ch {
		log.Println("======")
		log.Print(string(v.Data.ToByte()))
		log.Println("======")
	}
}
