# Week 4 demos

Eight programs and benchmarks for mutex internals, starvation, `RWMutex`, concurrent maps
and contention.

## Setup

- `go.mod` pins `go 1.27.1`. With `GOTOOLCHAIN=auto` (the default), any Go 1.21 or newer
  downloads 1.27.1 on first use, so you need network once.
- If your shell exports `GOROOT`, unset it. Otherwise the build can fail with
  `compile: version ... does not match go tool version`.
- Run every command from this directory, the module root. Shell scripts need a POSIX shell.

Build the four program demos first:

```sh
mkdir -p bin
for d in 03-statword 04-fairness 05-threshold 08-contention; do
    go build -o "bin/${d#*-}" "./$d"
done
```

Build first rather than using `go run`: a `GODEBUG` setting can otherwise leak into cached
build output.

## Run

In lecture order:

| Demo | Command | What it shows |
|---|---|---|
| 04-fairness | `GOMAXPROCS=2 ./bin/fairness` | `sync.Mutex`, an unfair test-and-set lock and a FIFO channel lock under contention: acquisition streaks and starvation gaps. |
| 02-sourcepath | `sh 02-sourcepath/run.sh` | Why the mutex algorithm is in `src/internal/sync/mutex.go`, not the small public wrapper in `src/sync/mutex.go`. |
| 01-fastpath | `sh 01-fastpath/dis.sh` | Inlining and the generated amd64 code for the mutex fast paths. The listing is architecture-dependent. |
| 01-fastpath | `go test -v -run TestSizes ./01-fastpath` | Sizes of `sync.Mutex` and other synchronization values on linux/amd64. |
| 01-fastpath | `go test -run=X -bench=. -count=5 -cpu=1,2 ./01-fastpath` | Uncontended mutex and `RWMutex` benchmarks. |
| 03-statword | `GOMAXPROCS=2 ./bin/statword` | A reconstructed state word: locked, woken and starving bits plus waiter accounting. The sampler itself is clean under `-race`. |
| 05-threshold | `GOMAXPROCS=2 ./bin/threshold` | Latency distributions around the one-millisecond starvation threshold. The threshold is not a deadline for `Lock` to return. |
| 06-rwmutex | `go test -run=X -bench=. -count=5 -cpu=2 ./06-rwmutex` | `Mutex` and `RWMutex` under several read/write mixes and critical-section workloads. |
| 07-maps | `GOMAXPROCS=2 go test -run=X -bench=. -count=5 -cpu=1,2 ./07-maps` | One global mutex, sharding and `sync.Map` under several operation mixes. |
| 08-contention | `GOMAXPROCS=2 ./bin/contention` | Mutex contention profiles and a packed-versus-padded layout measurement. |

The benchmark commands with `-count=5` can take several minutes. Run timing measurements
alone. Numbers depend on the machine and the invocation; compare like with like, not a
nanosecond value from one machine or week against another.

## The lecture cue runner

`sh cues.sh --list` prints every cue and runs nothing. `sh cues.sh` builds the programs and
walks through the cues one key at a time; `sh cues.sh D4` runs one cue.

## License

Apache-2.0, see [../../LICENSE-CODE](../../LICENSE-CODE).
