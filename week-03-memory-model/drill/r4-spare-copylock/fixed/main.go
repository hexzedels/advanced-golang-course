package main

import (
	"fmt"
	"sync"
)

type Status struct {
	mu   sync.Mutex
	last string
}

func (s *Status) Set(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.last = v
}

func main() {
	var s Status
	s.Set("ready")
	fmt.Printf("%q\n", s.last)
}
