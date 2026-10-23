package main

import "fmt"
import "sync/atomic"
import "time"

type Config struct{ Rev, Limit int } // Limit == Rev*10

var cfg atomic.Pointer[Config]

func handle() string {
	c := cfg.Load() // one snapshot per request
	return fmt.Sprintf("rev %d limit %d", c.Rev, c.Limit)
}

func main() {
	cfg.Store(&Config{1, 10})
	go func() { // hot reload
		for rev := 2; ; rev++ {
			cfg.Store(&Config{rev, rev * 10})
		}
	}()
	time.Sleep(time.Millisecond)
	fmt.Println(handle())
}
