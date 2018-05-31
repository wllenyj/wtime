package wtime

import (
	"time"
)

type Ticker struct {
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
	return t.r.w.resetTimer(t.r, d, d)
}

func (t *Ticker) Stop() bool {
	return t.r.w.delTimer(t.r)
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
