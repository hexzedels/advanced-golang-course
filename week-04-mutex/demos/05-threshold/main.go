// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
//
// Reconstructed 2026-10-25 from demos.md §5; not in the original document.
//
// Demo 5 (demos.md §5, lecture §6.4): what the 1 ms actually bounds.
//
// Sixteen greedy relockers plus one victim that times each of its own Lock
// calls. Three critical-section lengths, three locks (the tas and fifo types
// from demo 4, and sync.Mutex), 20,000 victim acquisitions each.
//
// demos.md gives the victim loop (used as given) and the output format.
// Inferred from the prose and the captured output:
//   - work is the spin(n) of demo 6, done inside the critical section by the
//     victim and by every greedy goroutine;
//   - greedy goroutines count acquisitions in batches of 100 and check the stop
//     flag once per batch, which is why the captured greedy ops are multiples
//     of 100; they stop when the victim has finished;
//   - percentiles are order statistics of the sorted waits: pN is
//     waits[N% of n], max is the largest, >1ms counts waits over 1 ms;
//   - the waits are printed as time.Duration, unrounded. The capture's
//     100 ns granularity is not produced by this program.
//
// Build it and run the binary.
package main

import (
	"fmt"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

const (
	greedy         = 16
	victimAcquires = 20000
)

var sink int

func spin(n int) {
	s := 0
	for i := 0; i < n; i++ {
		s += i ^ n
	}
	sink = s
}

type locker interface {
	Lock()
	Unlock()
}

// tas: pure barging, no queue at all (demo 4).
type unfair struct{ v atomic.Int32 }

func (u *unfair) Lock() {
	for {
		if u.v.Load() == 0 && u.v.CompareAndSwap(0, 1) {
			return
		}
		runtime.Gosched()
	}
}
func (u *unfair) Unlock() { u.v.Store(0) }

// fifo: a cap-1 channel; the chan sendq is FIFO (demo 4).
type fifo struct{ ch chan struct{} }

func newFifo() *fifo    { return &fifo{ch: make(chan struct{}, 1)} }
func (f *fifo) Lock()   { f.ch <- struct{}{} }
func (f *fifo) Unlock() { <-f.ch }

func run(name string, l locker, work int) {
	var (
		stop atomic.Bool
		ops  atomic.Int64
		done sync.WaitGroup
	)
	for range greedy {
		done.Add(1)
		go func() {
			defer done.Done()
			for !stop.Load() {
				for k := 0; k < 100; k++ {
					l.Lock()
					spin(work)
					l.Unlock()
				}
				ops.Add(100)
			}
		}()
	}

	waits := make([]time.Duration, 0, victimAcquires)
	time.Sleep(50 * time.Millisecond) // let the greedy goroutines saturate the lock
	for i := 0; i < victimAcquires; i++ {
		t0 := time.Now()
		l.Lock()
		w := time.Since(t0)
		spin(work)
		l.Unlock()
		waits = append(waits, w)
	}

	stop.Store(true)
	done.Wait()

	slices.Sort(waits)
	n := len(waits)
	at := func(p float64) time.Duration { return waits[int(p*float64(n))] }
	over := 0
	for _, w := range waits {
		if w > time.Millisecond {
			over++
		}
	}
	fmt.Printf("%-15s %6d %10v %10v %10v %10v %12v %8d %12d\n",
		name, n, at(0.50), at(0.90), at(0.99), at(0.999), waits[n-1], over, ops.Load())
}

func main() {
	fmt.Printf("GOMAXPROCS=%d NumCPU=%d, %d greedy relockers + 1 victim\n",
		runtime.GOMAXPROCS(0), runtime.NumCPU(), greedy)
	fmt.Println("starvationThresholdNs = 1e6  (src/internal/sync/mutex.go:55)")
	for _, work := range []int{0, 200, 2000} {
		fmt.Printf("\n--- critical section work=%d ---\n", work)
		fmt.Printf("%-15s %6s %10s %10s %10s %10s %12s %8s %12s\n",
			"lock", "n", "p50", "p90", "p99", "p99.9", "max", ">1ms", "greedy ops")
		run("tas", new(unfair), work)
		run("sync.Mutex", new(sync.Mutex), work)
		run("fifo chan", newFifo(), work)
	}
}
