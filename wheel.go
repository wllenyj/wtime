package wtime

import (
	//"container/list"
	"errors"
	//"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	TimerUsable   uint32 = 1
	TimerUnusable uint32 = 0

	defTimerSize = 128
)

var (
	Dration_ERR = errors.New("non-positive interval for wheel timer")

	t_bits = [...]uint8{8, 6, 6, 4, 4, 4}

	pool = &sync.Pool{
		New: func() interface{} {
			return make([]*timer, 0, defTimerSize)
		},
	}
)

type timeproc func(time.Time, interface{})

func sendTime(t time.Time, arg interface{}) {
	select {
	case arg.(chan time.Time) <- t:
	default:
	}
}

func GoFunc(t time.Time, arg interface{}) {
	go arg.(func())()
}

type timer struct {
	expires uint32
	period  uint32

	f   func(time.Time, interface{})
	arg interface{}

	state uint32

	w *Wheel
}

type Wheel struct {
	now atomic.Value

	add  chan *timer
	tran chan time.Time
	in   <-chan time.Time

	tick    time.Duration
	ticker  *time.Ticker
	jiffies uint32

	wheelers [6]*wheeler
	overflow chan time.Time

	quit   chan struct{}
	quited uint32

	wg sync.WaitGroup
}

func NewWheel(tick time.Duration) *Wheel {
	if tick <= 0 {
		panic(Dration_ERR)
	}
	w := &Wheel{
		add:    make(chan *timer, 1024),
		tick:   tick,
		ticker: time.NewTicker(tick),
		tran:   make(chan time.Time),

		overflow: make(chan time.Time, 1),

		quit: make(chan struct{}),
	}
	w.now.Store(time.Now())
	w.in = w.ticker.C

	w.wheelers[0] = &wheeler{
		i:    t_bits[0],
		j:    0,
		mask: 255,
		in:   w.tran,
		out:  w.overflow,
		add:  make(chan *timer, 512),
		w:    w,
		tv:   make([][]*timer, 1<<t_bits[0]),
		//tv:   make([]*list.List, 1<<t_bits[0]),
	}
	for i := range w.wheelers[0].tv {
		//w.wheelers[0].tv[i] = make([]*timer, 0, 256)
		//w.wheelers[0].tv[i] = make([]*timer, 0, 128)
		w.wheelers[0].tv[i] = pool.Get().([]*timer)
		//w.wheelers[0].tv[i] = list.New()
	}
	w.wg.Add(1)
	go w.wheelers[0].wheel()

	w.wg.Add(1)
	go w.loop()

	return w
}

func (w *Wheel) Stop() {
	atomic.StoreUint32(&w.quited, 1)
	close(w.quit)
	w.wg.Wait()
	w.ticker.Stop()
}

func (w *Wheel) newTimer(when time.Duration,
	period time.Duration,
	f timeproc,
	arg interface{}) *timer {

	if when < w.tick {
		panic(Dration_ERR)
	}
	if period != 0 && period < w.tick {
		panic(Dration_ERR)
	}

	t := &timer{
		expires: uint32(float64(when)/float64(w.tick) + 0.5),
		period:  uint32(float64(period)/float64(w.tick) + 0.5),
		//expires: uint32(when.Round(w.tick)/w.tick),
		//period:  uint32(period.Round(w.tick)/w.tick),
		f:     f,
		arg:   arg,
		state: TimerUsable,
		w:     w,
	}
	//fmt.Printf("%+v\n", t)
	return t
}

func (w *Wheel) newStopedTimer(f timeproc, arg interface{}) *timer {
	return &timer{
		f:     f,
		arg:   arg,
		state: TimerUnusable,
		w:     w,
	}
}

func (w *Wheel) addTimer(t *timer) {
	if atomic.LoadUint32(&w.quited) == 1 {
		return
	}
	w.add <- t
}

func (w *Wheel) delTimer(t *timer) bool {
	if atomic.LoadUint32(&w.quited) == 1 {
		return false
	}
	ret := atomic.CompareAndSwapUint32(&t.state, TimerUsable, TimerUnusable)
	return ret
}

func (w *Wheel) resetTimer(t *timer, when, period time.Duration) (*timer, bool) {
	if atomic.LoadUint32(&w.quited) == 1 {
		return t, false
	}
	deleted := atomic.CompareAndSwapUint32(&t.state, TimerUsable, TimerUnusable)
	//fmt.Printf("reset %v\n", ret)
	new_t := w.newTimer(when, period, t.f, t.arg)
	w.add <- new_t
	return new_t, deleted
}

func (w *Wheel) loop() {

	for {
		select {
		case t := <-w.in:
			w.jiffies++
			w.tran <- t
			w.now.Store(t)
		case t := <-w.add:
			w.addTimerInternal(t)
		case <-w.overflow:
			w.jiffies = 0
		case <-w.quit:
			//fmt.Printf("wheel quit\n")
			goto END_FOR
		}
	}
END_FOR:
	w.wg.Done()
}

func (w *Wheel) addTimerInternal(t *timer) {
	jiffies := w.jiffies

	for i := len(t_bits) - 1; i >= 0; i-- {
		rest := (32 - t_bits[i])
		if (t.expires >> rest) == 0 {
			jiffies <<= t_bits[i]
			t.expires <<= t_bits[i]
			//fmt.Printf("t %d expires:%08x ji:%08x\n", i, t.expires, jiffies)
		} else {
			w.createWheelers(i)
			//fmt.Printf("t create and add:%d expires:%08x ji:%08x\n", i, t.expires, jiffies)
			t.expires += jiffies & ^(uint32(w.wheelers[i].mask) << rest)
			//fmt.Printf("%08x += %08x & ^(%08x << %d)\n", t.expires, jiffies, w.wheelers[i].mask, rest)
			w.wheelers[i].add <- t
			return
		}
	}
}

func (w *Wheel) createWheelers(i int) {
	if w.wheelers[i] != nil {
		return
	}
	if w.wheelers[i-1] == nil {
		w.createWheelers(i - 1)
	}
	//overflow := w.wheelers[i-1].out
	w.wheelers[i-1].Lock()
	//bridge := make(chan time.Time)
	bridge := make(chan time.Time, 1)
	w.wheelers[i-1].out = bridge
	w.wheelers[i-1].Unlock()
	w.wheelers[i] = w.newWheeler(i, bridge, w.overflow, w.wheelers[i-1].add)
}

func (w *Wheel) newWheeler(i int, in chan time.Time, out chan<- time.Time, tran chan<- *timer) *wheeler {
	ret := &wheeler{
		i:    t_bits[i],
		j:    uint8(i),
		mask: (1 << t_bits[i]) - 1,
		in:   in,
		out:  out,
		add:  make(chan *timer, 64),
		//add:  make(chan *timer),
		tran: tran,
		w:    w,
		tv:   make([][]*timer, 1<<t_bits[i]),
		//tv:   make([]*list.List, 1<<t_bits[i]),
	}
	//fmt.Printf("new %d %d ji:%d cur:%d %b\n", ret.j, ret.i, w.jiffies, ret.curr, ret.mask)
	for i := range ret.tv {
		//ret.tv[i] = make([]*timer, 0, 128)
		//ret.tv[i] = list.New()
		ret.tv[i] = pool.Get().([]*timer)
	}
	//fmt.Printf("wheeler %+v\n", ret)
	w.wg.Add(1)
	go ret.wheel()
	return ret
}

type WheelDelegate interface {
	Done() <-chan struct{}
	addTimer(*timer)
	WaitDone()
}

func (w *Wheel) Done() <-chan struct{} {
	return w.quit
}

func (w *Wheel) WaitDone() {
	w.wg.Done()
}

type wheeler struct {
	sync.Mutex
	i    uint8
	j    uint8
	mask uint8
	curr uint8

	in  <-chan time.Time
	out chan<- time.Time

	add  chan *timer
	tran chan<- *timer

	tv [][]*timer
	//tv []*list.List

	w WheelDelegate
}

func (wr *wheeler) do(tm *time.Time) bool {
	select {
	case *tm = <-wr.in:
		//if wr.j >= 1 {
		//	fmt.Printf("tick %d  %d  %d %b\n", wr.curr, wr.i, wr.j, wr.mask)
		//}
		wr.curr++
		wr.curr &= wr.mask
		if wr.j == 0 {
			//fmt.Printf("%d-%d:tick curr:%d %08x\n", wr.i, wr.j, wr.curr, wr.mask)
		}
		if wr.curr == 0 {
			wr.Lock()
			wr.out <- *tm
			wr.Unlock()
		}
		vec := wr.tv[wr.curr]
		//if vec.Len() > 0 {
		//	//fmt.Printf("tick %d  %d  %d\n", wr.curr, wr.i, wr.j)
		//	front := vec.Front()
		//	vec.Init()
		//	go func() {
		//		for e := front; e != nil; e = e.Next() {
		//			t := e.Value.(*timer)
		//			wr.onTick(t, *tm)
		//		}
		//	}()
		//}
		if len(vec) > 0 {
			//wr.tv[wr.curr] = make([]*timer, 0, 128)
			wr.tv[wr.curr] = pool.Get().([]*timer)
			go func(tm time.Time) {
				for _, t := range vec {
					wr.onTick(t, tm)
				}
				vec = vec[:0:defTimerSize]
				pool.Put(vec)
			}(*tm)
		}
	case t := <-wr.add:
		//fmt.Printf("add expires:%08x %d\n", t.expires, wr.i)
		if t.expires != 0 {
			idx := uint8(t.expires >> (32 - wr.i))
			//fmt.Printf("push expires:%08x idx:%d curr:%d\n", t.expires, idx, wr.curr)
			t.expires <<= wr.i
			idx = (idx + wr.curr) & wr.mask
			//fmt.Printf("push expires:%08x index:%d\n", t.expires, idx+uint32(wr.curr))
			//wr.tv[idx].PushBack(t)
			wr.tv[idx] = append(wr.tv[idx], t)
			//t.f(*tm, t.arg)
			break
		}
		wr.onTick(t, *tm)
	case <-wr.w.Done():
		//fmt.Printf("wheeler quit\n")
		//goto END_FOR
		return false
	}
	return true
}

func (wr *wheeler) wheel() {
	var tm time.Time
	for wr.do(&tm) {
	}
	//fmt.Printf("wheeler real quit\n")
	wr.w.WaitDone()
}

func (wr *wheeler) onTick(t *timer, tm time.Time) {
	//fmt.Printf("onTick %08x \n", t.expires)
	if atomic.LoadUint32(&t.state) == TimerUnusable {
		//fmt.Printf("Unsable return %08x \n", t.expires)

		return
	}
	if t.expires == 0 {
		t.f(tm, t.arg)
		if t.period > 0 {
			t.expires = t.period
			wr.w.addTimer(t)
		} else {
			atomic.StoreUint32(&t.state, TimerUnusable)
		}
	} else {
		wr.tran <- t
	}
}
