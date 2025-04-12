package dict

import (
	"log"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestPut(t *testing.T) {
	dict := NewConcurrentDict(32768)
	var wg sync.WaitGroup
	wg.Add(300)
	t1 := time.Now()
	for i := 0; i < 300; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				dict.Put(strconv.Itoa(i)+"__"+strconv.Itoa(j), struct{}{})
			}
		}()
	}
	wg.Wait()
	log.Println(time.Since(t1).Milliseconds())
}

func putData(dict *ConcurrentDict) {
	for i := 0; i < 100000; i++ {
		dict.Put("key_"+strconv.Itoa(i), i)
	}
}

func TestGet(t *testing.T) {
	dict := NewConcurrentDict(8192)
	putData(dict)
	var wg sync.WaitGroup
	t1 := time.Now()
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100000; j++ {
				value, exists := dict.Get("key_" + strconv.Itoa(j))
				if !exists || value.(int) != j {
					t.Error("error!")
					return
				}
			}
		}()
	}
	wg.Wait()
	log.Println(time.Since(t1).Milliseconds())
}
