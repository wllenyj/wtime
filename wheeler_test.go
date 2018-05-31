package wtime

import (
	//"container/list"
	"sync"
	"testing"
	"time"
	//"fmt"
)

type testDelegate struct {
	w *Wheel
}

func (td *testDelegate) Done() <-chan struct{} {
	return td.w.Done()
}
func (td *testDelegate) addTimer(t *timer) {
	td.w.addTimerInternal(t)
}
func (td *testDelegate) WaitDone() {
	td.w.wg.Done()
}

func createWheel(d time.Duration, in <-chan time.Time) *Wheel {
	w := &Wheel{
		add:  make(chan *timer, 10),
		tick: d,

		tran:   make(chan time.Time, 1),
		in:     in,

		overflow: make(chan time.Time, 1),
		quit:     make(chan struct{}),
	}

	step1 := make(chan time.Time, 1)
	step2 := make(chan time.Time, 1)
	w.wheelers[0] = createTestWheeler(0, w, w.tran, step1, nil)
	w.wheelers[1] = createTestWheeler(1, w, step1, step2, w.wheelers[0].add)
	w.wheelers[2] = createTestWheeler(2, w, step2, w.overflow, w.wheelers[1].add)

	//w.wg.Add(1)
	//go w.wheelers[0].wheel()

	w.wg.Add(1)
	go w.loop()

	return w
}
func createTestWheeler(i int, w *Wheel, in <-chan time.Time, out chan time.Time, tran chan<- *timer) *wheeler {
	wr := &wheeler{
		i:    t_bits[i],
		mask: (1 << t_bits[0]) - 1,
		in:   in,
		out:  w.overflow,
		add:  make(chan *timer, 10),
		tran: tran,
		w:    &testDelegate{w: w},
		//tv:   make([]*list.List, 1<<t_bits[i]),
		tv:   make([][]*timer, 1<<t_bits[0]),
	}
	for i := range wr.tv {
		//wr.tv[i] = list.New()
		wr.tv[i] = make([]*timer, 0, defTimerSize)
	}
	return wr
}

func testFunc(tm time.Time, arg interface{}) {
	t := arg.(*testType)
	t.count++
	//fmt.Println("testfunc")
	if t.tm != tm {
    		t.t.Fatalf("test fail.")
	}
	t.wg.Done()
}

type testType struct {
	count int
	t     *testing.T
	tm    time.Time
	wg    sync.WaitGroup
}

func TestWheeler(t *testing.T) {
	in := make(chan time.Time, 1)
	test_t := time.Now()
	test_data := &testType{
		t:  t,
		tm: test_t,
	}
	var tt time.Time
	w := createWheel(1*time.Millisecond, in)
	test_data.wg.Add(1)
	tm := w.newTimer(10*time.Millisecond, 5*time.Millisecond, testFunc, test_data)
	w.addTimerInternal(tm)
	w.wheelers[0].do(&tt)
	if w.wheelers[0].curr != 0 {
		t.Fatalf("wheelers curr :%d", w.wheelers[0].curr)
	}
	//if w.wheelers[0].tv[10].Len() != 1 {
	if len(w.wheelers[0].tv[10]) != 1 {
		t.Fatalf("wheelers list 10 len:%d", len(w.wheelers[0].tv[10]))
	}
	for i := 0; i < 10; i++ {
		in <- test_t
		w.wheelers[0].do(&tt)
	}
	if w.wheelers[0].curr != 10 {
		t.Fatalf("wheelers curr :%d", w.wheelers[0].curr)
	}
	//if w.wheelers[0].tv[10].Len() != 0 {
	if len(w.wheelers[0].tv[10]) != 0 {
		t.Fatalf("wheelers list 10 len:%d", len(w.wheelers[0].tv[10]))
	}
	test_data.wg.Add(1)
	w.wheelers[0].do(&tt)
	//if w.wheelers[0].tv[15].Len() != 1 {
	if len(w.wheelers[0].tv[15]) != 1 {
		t.Fatalf("wheelers list 25 len:%d", len(w.wheelers[0].tv[15]))
	}
	for i := 0; i < 5; i++ {
		in <- test_t
		w.wheelers[0].do(&tt)
	}
	if w.wheelers[0].curr != 15 {
		t.Fatalf("wheelers curr :%d", w.wheelers[0].curr)
	}
	test_data.wg.Wait()
	if test_data.count != 2 {
		t.Fatalf("count %d", test_data.count)
	}
}

func TestMulWheelers(t *testing.T) {
	type runCount struct {
		i   int
		cnt int
	}
	test_data := []struct {
		d   int64
		p   int64
		i   int
		j   int
		len int
	}{
		{255, 0, 0, 255, 1},
		{256, 0, 1, 1, 1},
		{256, 0, 1, 1, 2},
		{0x10000, 0, 2, 4, 1},
	}

	in := make(chan time.Time, 1)
	test_t := time.Now()
	test_type := &testType{
		t:  t,
		tm: test_t,
	}
	var tt time.Time
	w := createWheel(1*time.Millisecond, in)
	var max int64 = 0
	for _, test := range test_data {
		if test.d > max {
			max = test.d
		}
		test_type.wg.Add(1)
		tm := w.newTimer(time.Duration(test.d)*time.Millisecond,
			time.Duration(test.p)*time.Millisecond, testFunc, test_type)
		w.addTimerInternal(tm)
		w.wheelers[test.i].do(&tt)
		i := test.i
		j := test.j
		//if w.wheelers[i].tv[j].Len() != test.len {
		if len(w.wheelers[i].tv[j]) != test.len {
			t.Fatalf("wheelers list [%d][%d] len:%d", i, j, len(w.wheelers[i].tv[j]))
		}
	}
}

func TestDelTimer(t *testing.T) {
	in := make(chan time.Time, 1)
	test_t := time.Now()
	test_data := &testType{
		t:  t,
		tm: test_t,
	}
	var tt time.Time
	w := createWheel(1*time.Millisecond, in)
	tm := w.newTimer(10*time.Millisecond, 0, testFunc, test_data)
	w.addTimerInternal(tm)
	test_data.wg.Add(1)
	w.wheelers[0].do(&tt)
	tm1 := w.newTimer(10*time.Millisecond, 0, testFunc, test_data)
	w.addTimerInternal(tm1)
	test_data.wg.Add(1)
	w.wheelers[0].do(&tt)
	tm2 := w.newTimer(10*time.Millisecond, 0, testFunc, test_data)
	w.addTimerInternal(tm2)
	test_data.wg.Add(1)
	w.wheelers[0].do(&tt)
	if w.wheelers[0].curr != 0 {
		t.Fatalf("wheelers curr :%d", w.wheelers[0].curr)
	}
	//if w.wheelers[0].tv[10].Len() != 3 {
	if len(w.wheelers[0].tv[10]) != 3 {
		t.Fatalf("wheelers list 10 len:%d", len(w.wheelers[0].tv[10]))
	}
	if w.delTimer(tm2) {
		test_data.wg.Done()
	} else {
		t.Fatal("del should ok")
	}
	for i := 0; i < 10; i++ {
		in <- test_t
		w.wheelers[0].do(&tt)
	}
	test_data.wg.Wait()
	if w.delTimer(tm1) {
		t.Fatal("del should false")
	}
	if w.wheelers[0].curr != 10 {
		t.Fatalf("wheelers curr :%d", w.wheelers[0].curr)
	}
	//if w.wheelers[0].tv[10].Len() != 0 {
	if len(w.wheelers[0].tv[10]) != 0 {
		t.Fatalf("wheelers list 10 len:%d", len(w.wheelers[0].tv[10]))
	}
	if test_data.count != 2 {
		t.Fatalf("count %d", test_data.count)
	}

}

