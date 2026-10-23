package main

import "sync"

var mu sync.Mutex
var ready bool
var codes map[string]int

func lookup(k string) int {
	if !ready { // fast path, no lock
		mu.Lock()
		if !ready {
			codes = map[string]int{"ru": 7}
			ready = true
		}
		mu.Unlock()
	}
	return codes[k]
}

func main() {
	done := make(chan int)
	go func() { done <- lookup("ru") }()
	println(lookup("ru"), <-done)
}
