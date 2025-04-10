package wait

import (
	"sync"
	"time"
)

// Wait is a wait group with timeout
type Wait struct {
	wait sync.WaitGroup
}

func (w *Wait) Add(delta int) {
	w.wait.Add(delta)
}

func (w *Wait) Done() {
	w.wait.Done()
}

func (w *Wait) Wait() {
	w.wait.Wait()
}

func (w *Wait) WaitWithTimeout(timeout time.Duration) {
	ch := make(chan struct{})
	go func() {
		w.wait.Wait()
		ch <- struct{}{}
	}()

	select {
	case <-ch:
	case <-time.After(timeout):
	}
}
