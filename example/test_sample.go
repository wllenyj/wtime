package main

import (
	"fmt"
	"github.com/wllenyj/wtime"
	"math"
	"math/rand"
	"os"
	//"os/signal"
	//"syscall"
	"time"
	"runtime/pprof"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	file, _ := os.Create("cpu.prof")
	pprof.StartCPUProfile(file)
	defer pprof.StopCPUProfile()

	rand.Seed(time.Now().UnixNano())
	w := wtime.NewWheel(200 * time.Millisecond)

	f := func(d int) {
		dura := time.Duration(d) * 200 * time.Millisecond
		//fmt.Printf("start dura: %s\n", dura)
		ticker := w.NewTicker(dura)
		var cnt uint
		for {
			cnt++
			//sys := time.NewTimer(dura + 900*time.Millisecond)
			start := time.Now()
			select {
			//case t := <-sys.C:
			//	fmt.Printf("[%s] %d timeout %s %s %s\n", dura, cnt, start, t, t.Sub(start))
			//	ticker.Stop()
			case t := <-ticker.C:
				end := time.Now()

				//if math.Abs(float64(end.Sub(start)-dura)) > float64(5000*time.Microsecond) {
				if math.Abs(float64(end.Sub(start)-dura)) > float64(700*time.Millisecond) {
					fmt.Printf("[%s] %d execout %s sub:%s\n", dura, cnt, end.Sub(t), end.Sub(start)-dura)
				}
				//sys.Stop()
			}
			if cnt >= 3 {
				ticker.Stop()
				wg.Done()
				return
			}
		}
	}
	//for i := 0; i < 500000; i++ {
	for i := 0; i < 300000; i++ {
		n := rand.Int31n(300)
		for n == 0 {
			n = rand.Int31n(300)
		}
		n += 50
		wg.Add(1)
		go f(int(n))
	}

	//c := make(chan os.Signal, 1)
	//signal.Notify(c,
	//	syscall.SIGINT,
	//)
	//for {
	//	s := <-c
	//	switch s {
	//	case syscall.SIGINT:
	//		goto END_FOR
	//	}
	//}
	//END_FOR:

	wg.Wait()
	w.Stop()
	fmt.Println("quit")
}
