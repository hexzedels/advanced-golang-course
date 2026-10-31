<!-- Prose: CC-BY-SA 4.0. -->

# Week 4 practice: mutex state, interviews and map design

Saturday 31 October 2026. One 60-minute session: source prediction, two interview rounds
and one design exercise. There is no practice module, live coding, benchmark ladder or
separate drill this week.

## Before the session

1. Sync your fork: click **Sync fork** on GitHub, then `git pull` (or
   `git pull upstream main`).
2. Check that `go version` prints `go1.27.1`. If it does not, see
   [Setup](../demos/README.md#setup).
3. Open `src/internal/sync/mutex.go` and `src/runtime/sema.go` in that toolchain's source
   tree. `go env GOROOT` prints the tree to use.

## In the session

- Predict the `Lock`/`lockSlow` and `Unlock`/`unlockSlow` state transitions, then follow a
  wakeup through `semrelease1`. Compare normal-mode competition with starvation-mode
  handoff.
- Answer interview questions about starvation, `RWMutex`, lock ordering, deadlock and
  livelock.
- Defend a concurrent map design: one global mutex, sharding or `sync.Map`. Start from the
  workload and the required invariants, then say how you would test the choice.

There is no new graded week-4 homework and nothing to submit for this session. Week 3's
`ShardedCounter` deadline was Friday 30 October at 23:59 MSK.
