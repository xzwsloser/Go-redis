package parse

import (
	"bytes"
	"log"
	"testing"
)

func TestParseMsgSingleLine(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte("+OK\r\n"))
	buf.Write([]byte(":100\r\n"))
	buf.Write([]byte("$4\r\nPING\r\n"))
	buf.Write([]byte("*3\r\n$3\r\nSET\r\n$3\r\nKEY\r\n$5\r\nVALUE\r\n"))
	ch := ParseStream(bytes.NewReader(buf.Bytes()))
	for v := range ch {
		if v.Error != nil {
			log.Print(v.Error.Error())
		}
		log.Println("======")
		log.Print(string(v.Data.ToByte()))
		log.Println("======")
	}
}
