// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
//
// Reconstructed 2026-10-25 from demos.md §3; not in the original document.
//
// Demo 3 (demos.md §3, lecture §6.1-6.2): watch a live sync.Mutex change modes.
//
// sync.Mutex is struct{ _ noCopy; mu internal/sync.Mutex } and
// internal/sync.Mutex is struct{ state int32; sema uint32 }, so the state word
// is the first four bytes of any sync.Mutex. A sampler goroutine reads it with
// atomic.LoadInt32 while workers contend for the lock, and counts the bits.
//
// demos.md gives the constants, state(), layoutCheck() and the sampler loop;
// those are used as given (the elided exit paths of layoutCheck written out).
// The rest is inferred from the prose and the captured output:
//   - the sweep is the eight (workers, work) rows of the ranges table, 400 ms
//     of sampling each, printed in demos.md's one-configuration format;
//   - a worker does Lock, spin(work), Unlock, the same spin as demo 6;
//   - workers count completed critical sections in batches of 100 and check
//     the stop flag once per batch, which is why the captured counts are
//     multiples of 100 (78000);
//   - a start barrier, so that no worker runs before the sampler starts
//     (docs/environment-notes.md §3, L-09).
//
// Build it and run the binary; never under -race (see demos.md §3).
package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	mutexLocked      = 1 << 0
	mutexWoken       = 1 << 1
	mutexStarving    = 1 << 2
	mutexWaiterShift = 3
)

const window = 400 * time.Millisecond

var sink int

func spin(n int) {
	s := 0
	for i := 0; i < n; i++ {
		s += i ^ n
	}
	sink = s
}

func state(m *sync.Mutex) int32 {
	return atomic.LoadInt32((*int32)(unsafe.Pointer(m)))
}

func layoutCheck() {
	if got := unsafe.Sizeof(sync.Mutex{}); got != 8 {
		fmt.Fprintf(os.Stderr, "unexpected sync.Mutex size %d, want 8; layout assumption broken\n", got)
		os.Exit(1)
	}
	var m sync.Mutex
	if s := state(&m); s != 0 {
		fmt.Fprintf(os.Stderr, "zero sync.Mutex has state %d, want 0; layout assumption broken\n", s)
		os.Exit(1)
	}
	m.Lock()
	if s := state(&m); s != mutexLocked {
		fmt.Fprintf(os.Stderr, "locked sync.Mutex has state %d, want %d; layout assumption broken\n", s, mutexLocked)
		os.Exit(1)
	}
	m.Unlock()
	fmt.Println("layout check: sizeof(sync.Mutex)=8, state word at offset 0, Lock sets bit 0  OK")
}

func run(workers, work int) {
	var (
		mu    sync.Mutex
		stop  atomic.Bool
		crit  atomic.Int64
		ready sync.WaitGroup
		done  sync.WaitGroup
	)
	gate := make(chan struct{})
	for range workers {
		ready.Add(1)
		done.Add(1)
		go func() {
			defer done.Done()
			ready.Done()
			<-gate
			for !stop.Load() {
				for k := 0; k < 100; k++ {
					mu.Lock()
					spin(work)
					mu.Unlock()
				}
				crit.Add(100)
			}
		}()
	}
	ready.Wait()
	close(gate)

	var samples, locked, woken, starving, enters int64
	var maxWaiters int32
	prevStarving := false
	deadline := time.Now().Add(window)
	func() {
		for {
			for k := 0; k < 500; k++ {
				s := state(&mu)
				samples++
				nowStarving := s&mutexStarving != 0
				if nowStarving {
					starving++
					if !prevStarving {
						enters++
					}
				}
				prevStarving = nowStarving
				if s&mutexLocked != 0 {
					locked++
				}
				if s&mutexWoken != 0 {
					woken++
				}
				if w := s >> mutexWaiterShift; w > maxWaiters {
					maxWaiters = w
				}
			}
			if time.Now().After(deadline) {
				return
			}
		}
	}()
	stop.Store(true)
	done.Wait()

	pct := func(n int64) float64 { return 100 * float64(n) / float64(samples) }
	fmt.Printf("%d workers, work=%d, %v: %d samples of the state word\n", workers, work, window, samples)
	fmt.Printf("    mutexLocked   set in %8.2f%% of samples\n", pct(locked))
	fmt.Printf("    mutexWoken    set in %8.2f%% of samples\n", pct(woken))
	fmt.Printf("    mutexStarving set in %8.2f%% of samples   <-- starvation mode\n", pct(starving))
	fmt.Printf("    observed transitions into starvation mode: %d (%.0f/s)\n", enters, float64(enters)/window.Seconds())
	fmt.Printf("    max waiter count observed: %d\n", maxWaiters)
	fmt.Printf("    critical sections completed: %d\n", crit.Load())
}

func main() {
	layoutCheck()
	fmt.Printf("GOMAXPROCS=%d NumCPU=%d\n", runtime.GOMAXPROCS(0), runtime.NumCPU())
	for _, c := range []struct{ workers, work int }{
		{2, 0}, {4, 0}, {8, 0}, {8, 100}, {8, 1000}, {8, 10000}, {64, 1000}, {64, 10000},
	} {
		fmt.Println()
		run(c.workers, c.work)
	}
}
