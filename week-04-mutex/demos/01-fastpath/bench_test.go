// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
//
// Demo 1 (demos.md §1, lecture §4.3): the uncontended fast paths.
//
// Methodology: every benchmark here loops `for range b.N`, not `for b.Loop()`.
// demos.md's copy used b.Loop, and the lecture runs this file at -cpu=1,2. At
// go1.27.1 a b.Loop benchmark under a multi-value -cpu list takes its first
// measurement at the GOMAXPROCS the previous -cpu entry left behind
// (docs/environment-notes.md §0.3, docs/corrections.md §1), and the two loop
// forms also disagree on short bodies at a single -cpu value. b.N is not
// affected by the first, so the -cpu=1,2 command in demos.md stays valid.
// Never compare a figure from this file with a b.Loop figure.

package fastpath

import (
	"sync"
	"testing"
)

var mu sync.Mutex
var rw sync.RWMutex
var n int

// Uncontended: this is the single-CAS fast path and nothing else.
func BenchmarkMutexUncontended(b *testing.B) {
	for range b.N {
		mu.Lock()
		mu.Unlock()
	}
}

func BenchmarkMutexUncontendedWork(b *testing.B) {
	for range b.N {
		mu.Lock()
		n++
		mu.Unlock()
	}
}

func BenchmarkTryLockUncontended(b *testing.B) {
	for range b.N {
		if mu.TryLock() {
			mu.Unlock()
		}
	}
}

func BenchmarkRWMutexRLockUncontended(b *testing.B) {
	for range b.N {
		rw.RLock()
		rw.RUnlock()
	}
}

func BenchmarkRWMutexLockUncontended(b *testing.B) {
	for range b.N {
		rw.Lock()
		rw.Unlock()
	}
}
