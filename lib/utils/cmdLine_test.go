package utils

import (
	"log"
	"testing"
)

func TestEquals(t *testing.T) {
	a := 1
	ap := &a
	bp := &a
	log.Println(ap == bp)
	log.Println("=====")
	log.Print(Equals(ap, bp))
}
