// L3-07 probe. Demo 4's case 04 (unbuffered) and case 05 (capacity 1), plus two variants
// of case 05: 05b adds a second send, 05s sleeps 1 ms between the send and the read.
// Each case runs n times in one process (05s at most 2000 times, it sleeps). The probe
// counts how often main read payload as 0, that is, before the goroutine's write.
//
//	go run . -n=100000          # counts, no race detector
//	go run -race . -case=05b    # one run under the race detector
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var payload, sink int

// once runs one case and returns the value main read.
func once(kind string) int {
	payload = 0
	c := make(chan int, 1)
	if kind == "04" {
		c = make(chan int)
	}
	done := make(chan struct{})
	go func() {
		payload = 1
		<-c
		close(done)
	}()
	c <- 0
	switch kind {
	case "05b":
		c <- 0 // the 2nd send completes only after the 1st receive (k+C rule, k=1, C=1)
	case "05s":
		time.Sleep(time.Millisecond) // later in time, but still no edge
	}
	v := payload
	<-done // orders this goroutine's write before the next round's payload = 0
	return v
}

func main() {
	n := flag.Int("n", 1, "rounds per case")
	which := flag.String("case", "all", "04, 05, 05b, 05s or all")
	flag.Parse()
	kinds := []string{"04", "05", "05b", "05s"}
	if *which != "all" {
		kinds = []string{*which}
	}
	for _, k := range kinds {
		switch k {
		case "04", "05", "05b", "05s":
		default:
			fmt.Fprintln(os.Stderr, "no such case:", k)
			os.Exit(2)
		}
		rounds := *n
		if k == "05s" {
			rounds = min(rounds, 2000)
		}
		zeros := 0
		for range rounds {
			v := once(k)
			sink += v
			if v == 0 {
				zeros++
			}
		}
		fmt.Printf("case %-3s rounds %d  read 0: %d (%.2f%%)\n", k, rounds, zeros, 100*float64(zeros)/float64(rounds))
	}
}
