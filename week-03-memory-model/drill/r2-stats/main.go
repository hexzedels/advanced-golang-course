package main

import "fmt"
import "sync"
import "sync/atomic"

var requests int64

func handle() { atomic.AddInt64(&requests, 1) }

func flush() (n int64) {
	n, requests = requests, 0 // plain read, plain reset
	return n
}

func main() {
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(handle)
	}
	first := flush()
	wg.Wait()
	fmt.Println(first + flush())
}
