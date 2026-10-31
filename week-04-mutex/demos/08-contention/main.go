// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
//
// Reconstructed 2026-10-25 from demos.md §8; not in the original document.
//
// Demo 8 (demos.md §8, lecture §10.1-10.2): where the time goes.
//
// Part 1: the mutex profile, 8 workers x 200,000 Lock/count++/Unlock with
// runtime.SetMutexProfileFraction(1). Part 2: the same workload with the lock
// and its counter padded onto separate cache lines, against the two packed
// together, three runs each.
//
// demos.md gives the profile-reading loop and the two layout types (used as
// given) and the output format. Inferred:
//   - BlockProfileRecord.Cycles for a mutex event is in CPU ticks
//     (src/runtime/sema.go:165-175 takes cputicks()), so the delay is
//     converted to nanoseconds with the cycles/second that runtime/pprof
//     prints in its debug=1 header: the only public source of that rate;
//   - "contention events counted" is the sum of Count over the records;
//   - wall clock runs from the start barrier to the last worker's exit, and
//     is printed rounded to the millisecond;
//   - in part 2 each run measures padded first, then packed, a fresh value of
//     each, with the profile off; ratio is packed/padded;
//   - a start barrier for the workers (docs/environment-notes.md §3, L-09);
//   - packed is heap-allocated (new), so its 16 bytes are 16-aligned and can
//     never straddle a cache line.
//
// Build it and run the binary.
package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"
)

const (
	workers = 8
	perG    = 200_000
)

// work runs workers x perG Lock / *count++ / Unlock and returns the wall clock.
func work(mu *sync.Mutex, count *int64) time.Duration {
	var ready, done sync.WaitGroup
	gate := make(chan struct{})
	for range workers {
		ready.Add(1)
		done.Add(1)
		go func() {
			defer done.Done()
			ready.Done()
			<-gate
			for range perG {
				mu.Lock()
				*count++
				mu.Unlock()
			}
		}()
	}
	ready.Wait()
	t0 := time.Now()
	close(gate)
	done.Wait()
	return time.Since(t0)
}

var cpsRE = regexp.MustCompile(`(?m)^cycles/second=(\d+)$`)

func cyclesPerSecond() float64 {
	var buf bytes.Buffer
	if err := pprof.Lookup("mutex").WriteTo(&buf, 1); err != nil {
		fmt.Fprintln(os.Stderr, "mutex profile:", err)
		os.Exit(1)
	}
	m := cpsRE.FindSubmatch(buf.Bytes())
	if m == nil {
		fmt.Fprintln(os.Stderr, "no cycles/second line in the mutex profile header")
		os.Exit(1)
	}
	v, _ := strconv.ParseFloat(string(m[1]), 64)
	return v
}

func part1() {
	runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(0)

	var mu sync.Mutex
	var count int64
	wall := work(&mu, &count)

	recs := make([]runtime.BlockProfileRecord, 0)
	for {
		n, ok := runtime.MutexProfile(recs)
		if ok {
			recs = recs[:n]
			break
		}
		recs = make([]runtime.BlockProfileRecord, n+16)
	}
	var events, cycles int64
	for _, r := range recs {
		events += r.Count
		cycles += r.Cycles
	}
	lost := float64(cycles) / cyclesPerSecond() * 1e9

	fmt.Printf("part 1  mutex profile, %d workers x %d acquisitions\n", workers, perG)
	fmt.Printf("        wall clock:                 %v\n", wall.Round(time.Millisecond))
	fmt.Printf("        profile entries:            %d (mutex)\n", len(recs))
	fmt.Printf("        contention events counted:  %d\n", events)
	fmt.Printf("        goroutine-nanoseconds lost: %.0f\n", lost)
	fmt.Printf("                                    (%.2f x wall clock)\n", lost/float64(wall.Nanoseconds()))
}

type padded struct {
	_     [64]byte
	mu    sync.Mutex
	_     [64]byte
	count int64
	_     [64]byte
}

type packed struct {
	mu    sync.Mutex
	count int64
}

func part2() {
	const runs = 3
	fmt.Printf("part 2  lock/counter layout, %d workers x %d acquisitions, %d runs each\n", workers, perG, runs)
	for i := 1; i <= runs; i++ {
		pd := new(padded)
		tPad := work(&pd.mu, &pd.count)
		pk := new(packed)
		tPack := work(&pk.mu, &pk.count)
		fmt.Printf("        run %d  padded %3dms   packed %3dms   ratio %.2f\n",
			i, tPad.Milliseconds(), tPack.Milliseconds(), float64(tPack)/float64(tPad))
	}
}

func main() {
	part1()
	fmt.Println()
	part2()
}
