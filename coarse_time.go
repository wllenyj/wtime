package wtime

import (
	"time"
)

func (w *Wheel) FNow() time.Time {
	t := w.now.Load().(time.Time)
	return t.Truncate(w.tick)
}

func (w *Wheel) CNow() time.Time {
	t := w.now.Load().(time.Time)	
	return t.Truncate(w.tick).Add(w.tick)
}
