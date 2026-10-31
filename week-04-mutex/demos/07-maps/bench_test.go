// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.
//
// Reconstructed 2026-10-25 from demos.md §7; not in the original document.
//
// Demo 7 (demos.md §7, lecture §9.3-9.4): sharded map vs sync.Map vs one
// global mutex. Five implementations of one tiny interface over a 4096-key
// space, pre-populated, hammered with RunParallel, at four write mixes.
//
// demos.md gives the store interface and the two shard types (used as given)
// and names the five implementations. Inferred from its tables: the sub-
// benchmark names (<mix>/<impl>), the mixes (read-only, 1 write in 100, in 10,
// in 2), a per-goroutine xorshift key stream, and a fresh pre-populated store
// per b.Run call with the timer reset after populating.
//
// The shard count is GOMAXPROCS(0)*8, read at package initialisation, before
// -cpu takes effect: that is why demos.md's -cpu=1 and -cpu=2 columns both say
// Sharded16 on a 2-CPU box. On another machine the name changes (Sharded160 at
// 20 CPUs); run with GOMAXPROCS=2 in the environment to get demos.md's 16.
//
// Methodology: RunParallel's pb.Next is driven by b.N, not b.Loop, so a
// multi-value -cpu list is safe (docs/environment-notes.md §0.3). Run the whole
// file as ONE go test invocation (-count=5 -cpu=1,2): the padded/unpadded
// comparison is only valid inside one invocation (demos.md, "The retraction").

package maps

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

const keys = 4096

type store interface {
	Load(k uint64) (uint64, bool)
	Store(k, v uint64)
}

// 1. GlobalMutex: one map under one sync.Mutex.
type globalMutex struct {
	mu sync.Mutex
	m  map[uint64]uint64
}

func (s *globalMutex) Load(k uint64) (uint64, bool) {
	s.mu.Lock()
	v, ok := s.m[k]
	s.mu.Unlock()
	return v, ok
}

func (s *globalMutex) Store(k, v uint64) {
	s.mu.Lock()
	s.m[k] = v
	s.mu.Unlock()
}

// 2. GlobalRWMutex: one map under one sync.RWMutex.
type globalRWMutex struct {
	mu sync.RWMutex
	m  map[uint64]uint64
}

func (s *globalRWMutex) Load(k uint64) (uint64, bool) {
	s.mu.RLock()
	v, ok := s.m[k]
	s.mu.RUnlock()
	return v, ok
}

func (s *globalRWMutex) Store(k, v uint64) {
	s.mu.Lock()
	s.m[k] = v
	s.mu.Unlock()
}

// 3. Sharded: GOMAXPROCS*8 shards, each a map with its own sync.Mutex.

// A shard header is a Mutex (8) plus a map pointer (8). Pad to exactly one
// cache line so that consecutive shards are CacheLinePadSize apart.
type shard struct {
	mu sync.Mutex
	m  map[uint64]uint64
	_  [64 - 16]byte
}

type sharded struct{ s []shard }

func (s *sharded) Load(k uint64) (uint64, bool) {
	sh := &s.s[k%uint64(len(s.s))]
	sh.mu.Lock()
	v, ok := sh.m[k]
	sh.mu.Unlock()
	return v, ok
}

func (s *sharded) Store(k, v uint64) {
	sh := &s.s[k%uint64(len(s.s))]
	sh.mu.Lock()
	sh.m[k] = v
	sh.mu.Unlock()
}

// 4. ShardedUnpadded.

// badShard is the same thing with the padding arithmetic wrong: it assumes an
// 8-byte header when the header is 16 bytes, so sizeof is 72 and consecutive
// shards straddle cache lines. This is the first version of this demo, kept
// because the difference is the point.
type badShard struct {
	mu sync.Mutex
	m  map[uint64]uint64
	_  [64 - 8]byte
}

type badSharded struct{ s []badShard }

func (s *badSharded) Load(k uint64) (uint64, bool) {
	sh := &s.s[k%uint64(len(s.s))]
	sh.mu.Lock()
	v, ok := sh.m[k]
	sh.mu.Unlock()
	return v, ok
}

func (s *badSharded) Store(k, v uint64) {
	sh := &s.s[k%uint64(len(s.s))]
	sh.mu.Lock()
	sh.m[k] = v
	sh.mu.Unlock()
}

// 5. SyncMap: sync.Map, which at 1.27 is internal/sync.HashTrieMap.
type syncMap struct{ m sync.Map }

func (s *syncMap) Load(k uint64) (uint64, bool) {
	v, ok := s.m.Load(k)
	if !ok {
		return 0, false
	}
	return v.(uint64), true
}

func (s *syncMap) Store(k, v uint64) { s.m.Store(k, v) }

// Read at package initialisation, before -cpu sets GOMAXPROCS.
var nShards = runtime.GOMAXPROCS(0) * 8

var impls = []struct {
	name string
	make func() store
}{
	{"GlobalMutex", func() store { return &globalMutex{m: make(map[uint64]uint64)} }},
	{"GlobalRWMutex", func() store { return &globalRWMutex{m: make(map[uint64]uint64)} }},
	{fmt.Sprintf("Sharded%d", nShards), func() store {
		s := &sharded{s: make([]shard, nShards)}
		for i := range s.s {
			s.s[i].m = make(map[uint64]uint64)
		}
		return s
	}},
	{fmt.Sprintf("Sharded%dUnpadded", nShards), func() store {
		s := &badSharded{s: make([]badShard, nShards)}
		for i := range s.s {
			s.s[i].m = make(map[uint64]uint64)
		}
		return s
	}},
	{"SyncMap", func() store { return new(syncMap) }},
}

var mixes = []struct {
	name       string
	writeEvery int // 1 write per N operations; 0 means read-only
}{
	{"readonly", 0},
	{"write1in100", 100},
	{"write1in10", 10},
	{"write1in2", 2},
}

var seed atomic.Uint64
var sink atomic.Uint64

func BenchmarkMaps(b *testing.B) {
	for _, m := range mixes {
		for _, im := range impls {
			b.Run(m.name+"/"+im.name, func(b *testing.B) {
				s := im.make()
				for k := range uint64(keys) {
					s.Store(k, k)
				}
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					x := seed.Add(0x9e3779b97f4a7c15) | 1
					i, acc := 0, uint64(0)
					for pb.Next() {
						x ^= x << 13
						x ^= x >> 7
						x ^= x << 17
						k := x % keys
						i++
						if m.writeEvery > 0 && i%m.writeEvery == 0 {
							s.Store(k, x)
						} else {
							v, _ := s.Load(k)
							acc += v
						}
					}
					sink.Add(acc)
				})
			})
		}
	}
}
