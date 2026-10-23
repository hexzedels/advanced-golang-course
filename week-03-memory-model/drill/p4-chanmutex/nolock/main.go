package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	var log []int
	for i := range 100 {
		wg.Go(func() {
			log = append(log, i)
		})
	}
	wg.Wait()
	fmt.Println(len(log))
}
