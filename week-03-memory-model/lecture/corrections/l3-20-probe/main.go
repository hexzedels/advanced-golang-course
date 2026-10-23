// Probe for L3-20: demo 7 (../../../demos/07-aba/main.go)
// with the same stack, the same workers and the same pop/push pairs, plus
// optional counters. With -count and -log off, push and pop are the demo's
// code plus a few flag checks.
package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	workers = flag.Int("workers", 4, "goroutines doing pop/push pairs")
	ops     = flag.Int("ops", 20_000, "pop/push pairs per worker")
	nodes   = flag.Int("nodes", 8, "nodes on the stack")
	mode    = flag.String("mode", "fresh", "fresh | recycled | pool")
	widen   = flag.Bool("widen", false, "runtime.Gosched between reading next and the CAS")
	count   = flag.Bool("count", false, "count ABA events (adds two atomic ops per pop)")
	runs    = flag.Int("runs", 1, "independent runs")
	logev   = flag.Bool("log", false, "log every stack event; print the first stale CAS with the events inside its window")
	logABA  = flag.Bool("logaba", false, "with -log: stop at the first ABA (stale or not) instead of the first stale CAS")
)

// Event log (-log): one global sequence, a ring of events, stopped at the
// first stale CAS. Node identity is val: recycled nodes keep their val.
type event struct {
	seq   uint64
	w     int
	what  string
	n, nx int // node, and the next it read or installed (-1 = nil)
}

const ringSize = 1 << 20

var (
	gseq    atomic.Uint64
	ring    [ringSize]event
	stopLog atomic.Bool
	staleAt struct {
		w         int
		read, cas uint64
		old, next int
		stale     bool
	}
)

func id(n *node) int {
	if n == nil {
		return -1
	}
	return n.val
}

func logE(w int, what string, n, nx *node) uint64 {
	if !*logev || stopLog.Load() {
		return 0
	}
	s := gseq.Add(1)
	ring[s%ringSize] = event{s, w, what, id(n), id(nx)}
	return s
}

type node struct {
	val    int
	next   *node
	pushes atomic.Uint64 // probe only: bumped by every push of this address
}

type stack struct{ top atomic.Pointer[node] }

type counters struct {
	aba      atomic.Int64 // CAS succeeded although old was popped and pushed again since the Load
	stale    atomic.Int64 // ... and the next it installed was no longer old.next
	selfloop atomic.Int64 // push found its own node on top: n.next = n
	retries  atomic.Int64 // failed CAS in pop
	loopRun  atomic.Bool  // this run had a self-loop push at any time, counted or not
	broken   atomic.Bool  // set by the first stale or selfloop; nothing is counted after it
	first    atomic.Value // "stale" or "selfloop": which broke the stack first
}

func (c *counters) on() bool { return *count && !c.broken.Load() }

func (c *counters) breakBy(what string) {
	if c.broken.CompareAndSwap(false, true) {
		c.first.Store(what)
	}
}

var c counters

func (s *stack) push(w int, n *node) {
	if *count || *logABA {
		n.pushes.Add(1)
	}
	for {
		old := s.top.Load()
		n.next = old
		if *count && old == n {
			c.loopRun.Store(true)
		}
		if c.on() && old == n {
			c.selfloop.Add(1)
			c.breakBy("selfloop")
		}
		if s.top.CompareAndSwap(old, n) {
			logE(w, "push", n, old)
			return
		}
	}
}

func (s *stack) pop(w int) *node {
	for {
		old := s.top.Load()
		if old == nil {
			return nil
		}
		var v uint64
		if *count || *logABA {
			v = old.pushes.Load()
		}
		next := old.next
		r := logE(w, "read", old, next)
		if *widen {
			runtime.Gosched()
		}
		if s.top.CompareAndSwap(old, next) {
			cs := logE(w, "pop", old, next)
			if *logev && r != 0 && cs != 0 && (old.next != next || (*logABA && old.pushes.Load() != v)) && stopLog.CompareAndSwap(false, true) {
				staleAt.w, staleAt.read, staleAt.cas, staleAt.old, staleAt.next = w, r, cs, id(old), id(next)
				staleAt.stale = old.next != next
			}
			if c.on() && old.pushes.Load() != v {
				c.aba.Add(1)
				if old.next != next {
					c.stale.Add(1)
					c.breakBy("stale")
				}
			}
			return old
		}
		logE(w, "fail", old, next)
		if c.on() {
			c.retries.Add(1)
		}
	}
}

type freelist struct {
	mu sync.Mutex
	ns []*node
}

func (f *freelist) put(n *node) { f.mu.Lock(); f.ns = append(f.ns, n); f.mu.Unlock() }
func (f *freelist) take() *node {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.ns) == 0 {
		return nil
	}
	n := f.ns[len(f.ns)-1]
	f.ns = f.ns[:len(f.ns)-1]
	return n
}

func walk(s *stack) (int, bool) {
	seen := make(map[*node]bool, *nodes)
	for p := s.top.Load(); p != nil; p = p.next {
		if seen[p] {
			return len(seen), true
		}
		seen[p] = true
	}
	return len(seen), false
}

func run() (lost int, cycle bool) {
	var s stack
	var f freelist
	pool := sync.Pool{New: func() any { return new(node) }}
	for i := range *nodes {
		s.push(-1, &node{val: i})
	}
	var wg sync.WaitGroup
	for w := range *workers {
		wg.Go(func() {
			for range *ops {
				n := s.pop(w)
				if n == nil {
					continue
				}
				switch *mode {
				case "recycled":
					f.put(n)
					m := f.take()
					if m == nil {
						m = &node{val: n.val}
					}
					s.push(w, m)
				case "pool":
					pool.Put(n)
					m := pool.Get().(*node)
					m.val = n.val
					s.push(w, m)
				default:
					s.push(w, &node{val: n.val})
				}
			}
		})
	}
	wg.Wait()
	cnt, cyc := walk(&s)
	return *nodes - cnt, cyc
}

func main() {
	flag.Parse()
	corrupt := 0
	hist := map[string]int{}
	firsts := map[string]int{}
	loopRuns := 0
	for range *runs {
		c.loopRun.Store(false)
		c.broken.Store(false)
		c.first = atomic.Value{}
		lost, cyc := run()
		if c.loopRun.Load() {
			loopRuns++
		}
		if f, ok := c.first.Load().(string); ok {
			firsts[f]++
		}
		if lost != 0 || cyc {
			corrupt++
		}
		hist[fmt.Sprintf("lost=%d,cycle=%v", lost, cyc)]++
	}
	fmt.Printf("mode=%s widen=%v count=%v workers=%d ops=%d nodes=%d GOMAXPROCS=%d runs=%d corrupt=%d hist=%v",
		*mode, *widen, *count, *workers, *ops, *nodes, runtime.GOMAXPROCS(0), *runs, corrupt, hist)
	if *count {
		fmt.Printf(" aba=%d stale=%d selfloop=%d retries=%d firstbreak=%v runs_with_selfloop=%d (counts stop at the first break in each run)",
			c.aba.Load(), c.stale.Load(), c.selfloop.Load(), c.retries.Load(), firsts, loopRuns)
	}
	fmt.Println()
	if *logev && stopLog.Load() {
		a := staleAt
		fmt.Printf("first logged CAS (stale=%v): ", staleAt.stale)
		fmt.Printf(" worker %d read old=%d next=%d at seq %d, CAS succeeded at seq %d (%d events between)\n",
			a.w, a.old, a.next, a.read, a.cas, a.cas-a.read-1)
		for q := a.read; q <= a.cas; q++ {
			e := ring[q%ringSize]
			if e.seq != q {
				fmt.Printf("  seq %d: overwritten\n", q)
				continue
			}
			fmt.Printf("  %6d w%d %-4s n=%d next=%d\n", e.seq-a.read, e.w, e.what, e.n, e.nx)
		}
	}
}
