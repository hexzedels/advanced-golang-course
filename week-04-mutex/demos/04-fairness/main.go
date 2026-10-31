// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
//
// Reconstructed 2026-10-25 from demos.md §4; not in the original document.
//
// Demo 4 (demos.md §4, lecture §2 cold open and §6.3): capture streaks and
// starve gaps for three locks with one interface.
//
// The methodology is BenchmarkMutexCapture (src/runtime/runtime_test.go:564-667),
// re-pointed at sync.Mutex. Each worker locks, bumps a shared atomic.Int64,
// unlocks, and looks at how far the counter moved since its own previous turn:
//
//	delta == 0  -> nobody else got in: a CAPTURE streak
//	delta == k  -> k other acquisitions happened while we waited: a STARVE gap
//
// demos.md gives the tas and fifo lockers (used as given) and the output
// format. Inferred from the prose and the captured output:
//   - streak90/99/Max and starve90/99/Max are percentiles over events, exact
//     (not the runtime benchmark's log2 buckets): one streak event per run of
//     delta == 0 turns, one starve event per turn with delta > 0. A worker's
//     first turn has no previous turn and is not counted; its last streak is.
//   - workers acquire in batches of 100 and check the stop flag once per
//     batch, which is why the captured acquire counts are multiples of 100;
//   - ns/op is wall clock from the start barrier to the last worker's exit,
//     divided by total acquisitions: throughput, not latency (demos.md gotcha);
//   - a start barrier (docs/environment-notes.md §3, L-09).
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
	workers = 8
	window  = 500 * time.Millisecond
)

type locker interface {
	Lock()
	Unlock()
}

// tas: pure barging, no queue at all.
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

// fifo: a cap-1 channel. The runtime's chan sendq is FIFO, so waiters are
// served in arrival order and nobody can barge.
type fifo struct{ ch chan struct{} }

func newFifo() *fifo    { return &fifo{ch: make(chan struct{}, 1)} }
func (f *fifo) Lock()   { f.ch <- struct{}{} }
func (f *fifo) Unlock() { <-f.ch }

// hist counts event lengths exactly: small values in an array, the rare large
// ones in a slice, so a worker never allocates per event.
type hist struct {
	small [1 << 16]int64
	large []int64
}

func (h *hist) add(v int64) {
	if v < int64(len(h.small)) {
		h.small[v]++
	} else {
		h.large = append(h.large, v)
	}
}

func (h *hist) merge(o *hist) {
	for i, c := range o.small {
		h.small[i] += c
	}
	h.large = append(h.large, o.large...)
}

// pct returns the smallest value v such that at least p of the events are <= v.
func (h *hist) pct(p float64) int64 {
	var n int64
	for _, c := range h.small {
		n += c
	}
	n += int64(len(h.large))
	if n == 0 {
		return 0
	}
	target := int64(p * float64(n))
	if float64(target) < p*float64(n) {
		target++
	}
	var seen int64
	for v, c := range h.small {
		seen += c
		if seen >= target {
			return int64(v)
		}
	}
	slices.Sort(h.large)
	return h.large[target-seen-1]
}

func (h *hist) max() int64 {
	if len(h.large) > 0 {
		return slices.Max(h.large)
	}
	for v := len(h.small) - 1; v >= 0; v-- {
		if h.small[v] > 0 {
			return int64(v)
		}
	}
	return 0
}

type result struct {
	acquires       int64
	streak, starve hist
	elapsed        time.Duration
}

func measure(l locker) *result {
	var (
		count atomic.Int64
		stop  atomic.Bool
		ready sync.WaitGroup
		done  sync.WaitGroup
	)
	gate := make(chan struct{})
	hs := make([]*[2]hist, workers)
	for w := range workers {
		hs[w] = new([2]hist)
		ready.Add(1)
		done.Add(1)
		go func(h *[2]hist) {
			defer done.Done()
			var prev, streak int64
			first := true
			ready.Done()
			<-gate
			for !stop.Load() {
				for k := 0; k < 100; k++ {
					l.Lock()
					ours := count.Add(1)
					l.Unlock()
					delta := ours - prev - 1
					prev = ours
					if first {
						first = false
						streak = 1
						continue
					}
					if delta == 0 {
						streak++
					} else {
						h[0].add(streak)
						h[1].add(delta)
						streak = 1
					}
				}
			}
			h[0].add(streak)
		}(hs[w])
	}
	ready.Wait()
	t0 := time.Now()
	close(gate)
	time.Sleep(window)
	stop.Store(true)
	done.Wait()
	r := &result{acquires: count.Load(), elapsed: time.Since(t0)}
	for _, h := range hs {
		r.streak.merge(&h[0])
		r.starve.merge(&h[1])
	}
	return r
}

func main() {
	fmt.Printf("GOMAXPROCS=%d, %d workers, %v each\n\n", runtime.GOMAXPROCS(0), workers, window)
	fmt.Printf("%-15s %11s %9s %9s %10s %9s %9s %10s %9s\n",
		"lock", "acquires", "streak90", "streak99", "streakMax", "starve90", "starve99", "starveMax", "ns/op")
	for _, c := range []struct {
		name string
		l    locker
	}{
		{"sync.Mutex", new(sync.Mutex)},
		{"unfair TAS", new(unfair)},
		{"fifo chan", newFifo()},
	} {
		r := measure(c.l)
		fmt.Printf("%-15s %11d %9d %9d %10d %9d %9d %10d %9.1f\n",
			c.name, r.acquires,
			r.streak.pct(0.90), r.streak.pct(0.99), r.streak.max(),
			r.starve.pct(0.90), r.starve.pct(0.99), r.starve.max(),
			float64(r.elapsed.Nanoseconds())/float64(r.acquires))
	}
	fmt.Println()
	fmt.Println("streak = consecutive acquisitions by one goroutine with nobody else getting in.")
	fmt.Println("starve = acquisitions by others between two turns of the same goroutine.")
}
