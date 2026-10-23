package main

import "fmt"

type pair struct{ a, b int }

var p pair

func main() {
	go func() {
		for i := 0; ; i++ {
			p = pair{i, i}
		}
	}()
	for n := 1; n <= 100_000_000; n++ {
		q := p
		if q.a != q.b {
			fmt.Println("read", q, "on read", n)
			return
		}
	}
	fmt.Println("no torn read")
}
