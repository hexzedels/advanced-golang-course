<!-- Prose: CC-BY-SA 4.0. -->

# Week 3 practice: data races and false sharing

Friday 23 October 2026, 19:30 MSK. There is no practice module this week. The first half
runs the lecture's demos in [../demos/](../demos/) on your own machine and ends with Tier 0
of the homework. The second half is an interview drill on the shared screen; its snippets
are in [../drill/](../drill/), to try after the session.

## Before the session

1. Sync your fork: click **Sync fork** on GitHub, then `git pull` (or
   `git pull upstream main`).
2. `cd week-03-memory-model/demos`. `go version` should print `go1.27.1`; if it does not,
   see [Setup](../demos/README.md#setup).

## First half: run these yourself

From `week-03-memory-model/demos`, when the session gets to each step:

```sh
go run ./01-counter                          # steps 1 and 2
GOMAXPROCS=1 go run ./01-counter
GOMAXPROCS=1 go run -race ./01-counter       # right answer, still flagged
go run -race ./01-counter
go test -run=X -bench=. -benchtime=3x -count=5 -cpu=1 ./01-counter    # step 4, one -cpu per run
go test -run=X -bench=. -benchtime=3x -count=5 -cpu=2 ./01-counter
go test -run=X -bench=. -benchtime=3x -count=5 -cpu=8 ./01-counter
GOMAXPROCS=1 go run ./05-falsesharing/stats  # step 5
go run ./05-falsesharing/stats
go test -run=X -bench=. -benchtime=20x -count=5 -cpu=2 ./05-falsesharing   # step 6
```

- Each `-race` run waits about a second before it exits. That is the race detector, not a
  hang. `exit status 66` is the program's code for "race found".
- On Linux and Windows `-race` needs cgo, so a C compiler. If it will not build
  (`go: -race requires cgo`), follow the shared screen. Windows on arm64 has no race
  detector.
- **Windows PowerShell** has no `VAR=value command` form:
  `$env:GOMAXPROCS=1; go run ./01-counter; Remove-Item Env:GOMAXPROCS`.
- Do not run `03-reorder` under `-race`: it reports its races and then appears to hang.
- Numbers depend on the machine. Paste yours into the chat; compare ratios, not
  nanoseconds.

## Tier 0 of `ShardedCounter`

In [../homework/](../homework/), write "Tier 0" as its README describes: one `sync.Mutex`
around one `int64`. Edit `shardedcounter.go`: replace the three `panic` stubs in `New`,
`Add` and `Sum`, give the struct a `sync.Mutex` and an `int64` field, and import `sync`.
Then:

```sh
cd ../homework
go vet .
go test -race -count=20 -cpu=1,2,8 .
```

When that prints `ok`, measure it against the grader's copy of Tier 0:

```sh
go test -run='^$' -bench=AddParallel -count=5 -cpu=8 .
```

The two rows come out about equal: right now they are the same code. The homework asks for
more; its README says how much and how it is measured.
