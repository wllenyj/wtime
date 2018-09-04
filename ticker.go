package wtime

import (
	"time"
	"sync"
)

type Ticker struct {
	sync.RWMutex
	C <-chan time.Time
	r *timer
}

func (w *Wheel) NewTicker(d time.Duration) *Ticker {
	c := make(chan time.Time, 1)
	t := &Ticker{
		C: c,
		r: w.newTimer(d, d, sendTime, c),
	}
	w.addTimer(t.r)
	return t
}

func (t *Ticker) Reset(d time.Duration) bool {
	t.Lock()
	new, ret := t.r.w.resetTimer(t.r, d, d)
	t.r = new
	t.Unlock()

	return ret
}

func (t *Ticker) Stop() bool {
	t.RLock()
	ret := t.r.w.delTimer(t.r)
	t.RUnlock()
	return ret
}

func (w *Wheel) Tick(d time.Duration) <-chan time.Time {
	return w.NewTicker(d).C
}

func (w *Wheel) TickFunc(d time.Duration, f func()) *Ticker {
	t := &Ticker{
		r: w.newTimer(d, d, GoFunc, f),
	}
	w.addTimer(t.r)
	return t
}
