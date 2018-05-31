package main

import (
	"fmt"
	"github.com/wllenyj/wtime"
	"time"
)

func main() {
	w := wtime.NewWheel(1000 * time.Millisecond)

	for i:=0; i<100; i++{
		time.Sleep(1000*time.Millisecond)
		fmt.Printf("\n %s\n", time.Now())
		fmt.Println("floor ", w.FloorNow())
		fmt.Println("ceili ", w.CeilingNow())
	}

}
