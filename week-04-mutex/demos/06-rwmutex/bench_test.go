// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
//
// Reconstructed 2026-10-25 from demos.md §6; not in the original document.
//
// Demo 6 (demos.md §6, lecture §8.2): when RWMutex actually wins.
//
// RunParallel over a read-mostly workload with a tunable amount of
// deterministic work inside the critical section, against a sync.Mutex and a
// sync.RWMutex behind the same interface. demos.md gives spin() and bench();
// the guard interface, the two guards and the sub-benchmark grid
// (work 0, 10, 100, 1000 x readonly, 1 write in 100, 1 write in 10) are
// inferred from its table.
//
// One deliberate change to the given code. demos.md's spin ends with
// `sink = s`, a store to one package-level int. Two readers holding RLock run
// it at the same time, so it is a data race (`go test -race` reports a write
// against a write in spin), and it puts a contended store to a shared cache
// line inside every read critical section, which is not the cost this demo
// says it measures. Here spin returns s; each RunParallel goroutine adds the
// results into a local and publishes once, after its loop, so the loop cannot
// be optimised away and the critical section touches no shared memory.
//
// Methodology: RunParallel's pb.Next is driven by b.N, not b.Loop, so the
// b.Loop/-cpu trap (docs/environment-notes.md §0.3) does not apply; the
// lecture runs this file at the single value -cpu=2, where RunParallel starts
// GOMAXPROCS = 2 goroutines.

package rwmutex

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

var sink atomic.Int64

func spin(n int) int {
	s := 0
	for i := 0; i < n; i++ {
		s += i ^ n
	}
	return s
}

type guard interface {
	lock()
	unlock()
	rlock()
	runlock()
}

// mutexGuard: for a plain Mutex, rlock and lock are the same method.
type mutexGuard struct{ mu sync.Mutex }

func (g *mutexGuard) lock()    { g.mu.Lock() }
func (g *mutexGuard) unlock()  { g.mu.Unlock() }
func (g *mutexGuard) rlock()   { g.mu.Lock() }
func (g *mutexGuard) runlock() { g.mu.Unlock() }

type rwGuard struct{ mu sync.RWMutex }

func (g *rwGuard) lock()    { g.mu.Lock() }
func (g *rwGuard) unlock()  { g.mu.Unlock() }
func (g *rwGuard) rlock()   { g.mu.RLock() }
func (g *rwGuard) runlock() { g.mu.RUnlock() }

// writeEvery: 1 write per N operations. 0 means read-only.
func bench(b *testing.B, g guard, work int, writeEvery int) {
	b.RunParallel(func(pb *testing.PB) {
		i, acc := 0, 0
		for pb.Next() {
			i++
			if writeEvery > 0 && i%writeEvery == 0 {
				g.lock()
				acc += spin(work)
				g.unlock()
			} else {
				g.rlock()
				acc += spin(work)
				g.runlock()
			}
		}
		sink.Add(int64(acc))
	})
}

var mixes = []struct {
	name       string
	writeEvery int
}{
	{"readonly", 0},
	{"write1in100", 100},
	{"write1in10", 10},
}

func grid(b *testing.B, newGuard func() guard) {
	for _, work := range []int{0, 10, 100, 1000} {
		for _, m := range mixes {
			b.Run(fmt.Sprintf("work%d/%s", work, m.name), func(b *testing.B) {
				bench(b, newGuard(), work, m.writeEvery)
			})
		}
	}
}

func BenchmarkMutex(b *testing.B)   { grid(b, func() guard { return new(mutexGuard) }) }
func BenchmarkRWMutex(b *testing.B) { grid(b, func() guard { return new(rwGuard) }) }
