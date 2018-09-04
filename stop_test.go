package wtime

import (
	"sync"
	"testing"
	"time"
)

func TestStop(t *testing.T) {
	w := NewWheel(200 * time.Millisecond)
	pool := sync.Pool{
		New: func() interface{} {
			timer := w.NewStopedTimer()
			return timer
		},
	}

	poolGet := func() *Timer {
		t := pool.Get().(*Timer)
		return t
	}
	poolPut := func(tm *Timer) {
		tm.Stop()
		select {
		case <-tm.C:
		default:
		}
		pool.Put(tm)
	}
	ticker := time.NewTicker(1 * time.Second)
	done := time.NewTimer(5 * time.Second)
	for {
		testTimer := poolGet()
		testTimer.Reset(2 * time.Second)
		select {
		case <-ticker.C:
		case <-testTimer.C:
			t.Fatalf("test Timer failed")
		case <- done.C:
			return
		}
		poolPut(testTimer)
	}
}
