package wtime

import (
	"sync"
	"time"
)

type Timer struct {
	sync.RWMutex
	C <-chan time.Time
	r *timer
}

func (w *Wheel) NewTimer(d time.Duration) *Timer {
	c := make(chan time.Time, 1)
	t := &Timer{
		C: c,
		r: w.newTimer(d, 0, sendTime, c),
	}
	w.addTimer(t.r)
	return t
}

func (w *Wheel) NewStopedTimer() *Timer {
	c := make(chan time.Time, 1)
	t := &Timer{
		C: c,
		r: w.newStopedTimer(sendTime, c),
	}
	return t
}

func (t *Timer) Reset(d time.Duration) bool {
	t.Lock()
	new, ret := t.r.w.resetTimer(t.r, d, 0)
	t.r = new
	t.Unlock()
	return ret
}

func (t *Timer) Stop() bool {
	t.RLock()
	ret := t.r.w.delTimer(t.r)
	t.RUnlock()
	return ret
}

func (w *Wheel) After(d time.Duration) <-chan time.Time {
	return w.NewTimer(d).C
}

func (w *Wheel) Sleep(d time.Duration) {
	<-w.NewTimer(d).C
}

func (w *Wheel) AfterFunc(d time.Duration, f func()) *Timer {
	t := &Timer{
		r: w.newTimer(d, 0, GoFunc, f),
	}
	w.addTimer(t.r)
	return t
}
