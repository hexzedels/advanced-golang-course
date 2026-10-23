package main

import "fmt"
import "sync"
import "sync/atomic"

var requests atomic.Int64

func handle() { requests.Add(1) }

func flush() int64 { return requests.Swap(0) }

func main() {
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(handle)
	}
	first := flush()
	wg.Wait()
	fmt.Println(first + flush())
}
