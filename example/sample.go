package main

import (
	"fmt"
	"github.com/wllenyj/wtime"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	w := wtime.NewWheel(500 * time.Millisecond)

	f := func(d int) {
		dura := time.Duration(d) * 500 * time.Millisecond
		//fmt.Printf("start dura: %s\n", dura)
		ticker := w.NewTicker(dura)
		var cnt uint
		for {
			cnt++
			start := time.Now()
			select {
			case t := <-ticker.C:
				end := time.Now()

				//if math.Abs(float64(end.Sub(start)-dura)) > float64(5000*time.Microsecond) {
				if math.Abs(float64(end.Sub(start)-dura)) > float64(800*time.Millisecond) {
					fmt.Printf("[%s] %d execout %s %s - %s sub:%s\n", dura, cnt, t, end, start, end.Sub(start)-dura)
				}
			}
		}
	}
	for i := 0; i < 500000; i++ {
		n := rand.Int31n(30)
		for ; n == 0; {
			n = rand.Int31n(30)
		}
		n += 20
		go f(int(n))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGINT,
	)
	for {
		s := <-c
		switch s {
		case syscall.SIGINT:
			goto END_FOR
		}
	}
END_FOR:

	w.Stop()
	fmt.Println("quit")
}
