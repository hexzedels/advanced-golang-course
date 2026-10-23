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
			sq := i * i
			wg.Add(1)
			defer wg.Done()
			done <- sq
		}()
	}
	wg.Wait()
	fmt.Println(len(done))
}
