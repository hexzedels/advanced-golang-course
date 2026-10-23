package main

import (
	"fmt"
	"time"
)

var stop bool

func main() {
	go func() {
		time.Sleep(time.Millisecond)
		stop = true
	}()
	for !stop {
	}
	fmt.Println("loop ended")
}
