package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	done := make(chan int, 10)
	for i := range 10 {
		go func() {
			wg.Add(1)
			defer wg.Done()
			done <- i
		}()
	}
	wg.Wait()
	fmt.Println(len(done))
}
