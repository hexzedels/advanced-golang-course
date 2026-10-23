//go:build go1.21

package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Print(i, " ")
		}()
	}
	wg.Wait()
	fmt.Println()
}
