package main

import (
	"fmt"
	"sync/atomic"
)

var msg string
var ok atomic.Bool

func main() {
	go func() {
		msg = "hello"
		ok.Store(true)
	}()
	for !ok.Load() {
	}
	fmt.Printf("%q\n", msg)
}
