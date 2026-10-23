<!-- Prose: CC-BY-SA 4.0. -->

# Week 3 interview drill

The snippets from the second half of the practice on 23 October. Two rounds, the way an
interviewer asks them: "what does it print, and is that guaranteed?", then "review this
code like a PR".

Try each one before you run it. A wrong answer that a run happened to confirm is the
point of the drill: a racy program can print the right number.

## Setup

As in [../demos/](../demos/README.md#setup): `go.mod` pins `go 1.27.1`, unset `GOROOT` if
your shell exports it, and run every command from this directory.

Three directories fail `go vet` on purpose (`p1-loopvar/go121`, `p3-addinside`,
`r4-spare-copylock`): the warning is the bug. Run `go vet` on one directory at a time,
not on `./...`.

`go run -race` prints `exit status 66` when the race detector found a race, and it waits
about a second before it exits.

## Round 1: predict

| Snippet | Question | Also try |
|---|---|---|
| `p1-loopvar` | Goroutines in a loop print `i`. What comes out? | `go121/`, the same code built as Go 1.21 |
| `p2-publish` | Publish a string through an `atomic.Bool` flag. Is the string always there? | `swapped/`, the two stores in the other order |
| `p3-addinside` | `wg.Add` inside the goroutine. What does `Wait` wait for? | `vetblind/`, a shape `go vet` does not catch |
| `p4-chanmutex` | A buffered channel of one used as a lock. Is the slice safe? | `nolock/`, the same code without the channel |
| `p5-spinflag` | Spin on a plain `bool` until another goroutine sets it. | |
| `p6-spare-tearing` | Read a two-field struct while another goroutine writes it. Torn read? | `forever/`, an endless writer |

## Round 2: review and fix

Each has a `fixed/` directory. Find the bug before you open it.

| Snippet | What it is |
|---|---|
| `r1-lazyinit` | Lazy init with a lock-free fast path |
| `r2-stats` | A stats counter, half of it atomic |
| `r3-config` | A hot-reloaded config |
| `r4-spare-copylock` | A mutex in a struct with a value receiver |

Run a snippet as `go run ./r1-lazyinit`, then `go run -race ./r1-lazyinit`, then the same
for `./r1-lazyinit/fixed`.
