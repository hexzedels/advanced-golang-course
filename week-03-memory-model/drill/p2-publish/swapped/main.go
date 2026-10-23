package main

import (
	"fmt"
	"sync/atomic"
)

var msg string
var ok atomic.Bool

func main() {
	go func() {
		ok.Store(true)
		msg = "hello"
	}()
	for !ok.Load() {
	}
	fmt.Printf("%q\n", msg)
}
