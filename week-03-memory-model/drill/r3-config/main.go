package main

import "fmt"
import "time"

type Config struct{ Rev, Limit int } // Limit == Rev*10

var cfg = &Config{1, 10}

func handle() string {
	return fmt.Sprintf("rev %d limit %d", cfg.Rev, cfg.Limit)
}

func main() {
	go func() { // hot reload
		for rev := 2; ; rev++ {
			cfg = &Config{rev, rev * 10}
		}
	}()
	time.Sleep(time.Millisecond)
	fmt.Println(handle())
}
