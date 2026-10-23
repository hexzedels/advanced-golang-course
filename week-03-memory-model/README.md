# Week 3: The memory model

What the Go memory model promises and what it leaves out: data races, happens-before and
every edge that creates one, what the compiler and the hardware may reorder, torn
multi-word values, what an atomic costs, cache lines and false sharing, and where ABA
lives in Go.

The runnable demos are in [demos/](demos/). The practice session, the race detector on
your own machine, the cost of the three fixes and false sharing, is in
[practice/](practice/), and the interview drill from its second half is in [drill/](drill/).
The homework, `ShardedCounter`, is in [homework/](homework/), due Friday 30 October, 23:59
MSK. Corrections to the lecture are in [lecture/corrections/](lecture/corrections/). The
slides come later.
