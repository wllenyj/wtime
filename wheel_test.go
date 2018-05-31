package wtime

import (
	//"container/list"
	//"sync"
	//"fmt"
	//"runtime"
	"testing"
	"time"
)

//func NewWheelTest(tick time.Duration, in <-chan time.Time) *Wheel {
//	w := &Wheel{
//		add:    make(chan *timer),
//		tick:   tick,
//		ticker: time.NewTicker(10 * time.Minute),
//		tran:   make(chan time.Time),
//		in:     in,
//
//		overflow: make(chan time.Time, 1),
//
//		quit: make(chan struct{}),
//	}
//
//	w.wheelers[0] = &wheeler{
//		i:    t_bits[0],
//		mask: (1 << t_bits[0]) - 1,
//		in:   w.tran,
//		out:  w.overflow,
//		add:  make(chan *timer),
//		w:    w,
//		tv:   make([]*list.List, 1<<t_bits[0]),
//	}
//	for i := range w.wheelers[0].tv {
//		w.wheelers[0].tv[i] = list.New()
//	}
//	w.wg.Add(1)
//	go w.wheelers[0].wheel()
//
//	w.wg.Add(1)
//	go w.loop()
//
//	return w
//}
//
//func TestWheel(t *testing.T) {
//	in := make(chan time.Time)
//	//var tt time.Time
//	w := NewWheelTest(1*time.Millisecond, in)
//
//	test_arr := []struct {
//		d   time.Duration
//		cnt int
//		res int
//	}{
//		{0x10, 1, 1},
//		{0x100, 2, 1},
//		{0x4000, 3, 1},
//		{0x43450, 3, 1},
//	}
//
//	for _, test := range test_arr {
//		test_t := time.Now()
//		test_data := &testType{
//			t:  t,
//			tm: test_t,
//		}
//		test_data.wg.Add(1)
//		tm := w.newTimer(test.d*time.Millisecond, 0, testFunc, test_data)
//		fmt.Printf("for begin %08x\n", int64(test.d))
//		w.addTimer(tm)
//		//test_data.wg.Wait()
//		//for i := 0; i < test.cnt; i++ {
//		//	test_data.wg.Add(1)
//		//}
//		i := 0
//		for {
//			if i < int(test.d) {
//				in <- test_t
//			} else {
//				break
//			} 
//			if i%256 == 0 {
//				runtime.Gosched()
//				//time.Sleep(400 * time.Millisecond)
//			}
//			i++
//		}
//		fmt.Println("for end")
//		test_data.wg.Wait()
//		if test_data.count != test.res {
//			t.Fatalf("count %d", test_data.count)
//		}
//	}
//	w.Stop()
//}

func TestSleep(t *testing.T) {
	test_data := []struct {
		d   time.Duration
		sub time.Duration
	}{
		{200, 200},
		{300, 400},
		{299, 200},
	}
	w := NewWheel(200 * time.Millisecond)
	for _, data := range test_data {
		start := time.Now()
		//t.Logf("%s", start)
		w.Sleep(data.d * time.Millisecond)
		end := time.Now()
		t.Logf("sub:%s", end.Sub(start)-data.sub*time.Millisecond)
		if end.Sub(start)-data.sub*time.Millisecond > time.Millisecond {
			t.Fatalf("sleep fatal. %s %s", start, end)
		}
	}
	start := time.Now()
	time.Sleep(200 * time.Millisecond)
	end := time.Now()
	t.Logf("go timer sub:%s", end.Sub(start)-200*time.Millisecond)
	if end.Sub(start)-200*time.Millisecond > time.Millisecond {
		t.Fatalf("sleep fatal. %s %s", start, end)
	}
	w.Stop()
}
