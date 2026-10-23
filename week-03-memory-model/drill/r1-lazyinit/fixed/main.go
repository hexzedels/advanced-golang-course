package main

import "sync"

var codes = sync.OnceValue(func() map[string]int {
	return map[string]int{"ru": 7}
})

func lookup(k string) int { return codes()[k] }

func main() {
	done := make(chan int)
	go func() { done <- lookup("ru") }()
	println(lookup("ru"), <-done)
}
