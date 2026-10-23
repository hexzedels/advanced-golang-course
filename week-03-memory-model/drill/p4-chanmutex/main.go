package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 1)
	var log []int
	for i := range 100 {
		wg.Go(func() {
			sem <- struct{}{}
			log = append(log, i)
			<-sem
		})
	}
	wg.Wait()
	fmt.Println(len(log))
}
