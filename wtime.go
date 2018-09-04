package wtime

import (
	"time"
)

var defWheel *Wheel

func init() {
	defWheel = NewWheel(500 * time.Millisecond)
}

func NewTimer(d time.Duration) *Timer {
	return defWheel.NewTimer(d)
}

func NewStopedTimer() *Timer {
	return defWheel.NewStopedTimer()
}

func Sleep(d time.Duration) {
	defWheel.Sleep(d)
}

func After(d time.Duration) <-chan time.Time {
	return defWheel.After(d)
}

func AfterFunc(d time.Duration, f func()) *Timer {
	return defWheel.AfterFunc(d, f)
}

func NewTicker(d time.Duration) *Ticker {
	return defWheel.NewTicker(d)
}

func Tick(d time.Duration) <-chan time.Time {
	return defWheel.Tick(d)
}

func TickFunc(d time.Duration, f func()) *Ticker {
	return defWheel.TickFunc(d, f)
}

func FNow() time.Time {
	return defWheel.FNow()
}

func CNow() time.Time {
	return defWheel.CNow()
}
