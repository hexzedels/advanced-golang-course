# Advanced Golang

A free nine-week community course on concurrency in Go, from the scheduler to lock-free
data structures.

Lectures are live, in Russian, on Wednesdays; practice sessions are on Fridays. Both start
at 19:30 MSK. The course runs from 7 October to 8 December 2026.

## Weeks

1. [Goroutines and the scheduler](week-01-scheduler/)
2. [Channels](week-02-channels/)
3. [The memory model](week-03-memory-model/)
4. [Mutexes and the runtime semaphore](week-04-mutex/)
5. The sync toolbox
6. Context
7. Concurrency patterns
8. Testing concurrent code
9. Lock-free data structures

Material is added week by week.

## Homework

Fork this repository and solve each week's homework in your fork, under
`week-NN-*/homework/`. Send the link to your fork to the course chat once. The first task
and the full rules are in [week-01-scheduler/homework/](week-01-scheduler/homework/).

## Toolchain

Every module pins `go 1.27.1`. See [week-01-scheduler/demos/](week-01-scheduler/demos/)
for how that works and how to run the demos.

## License

Code is Apache-2.0 ([LICENSE-CODE](LICENSE-CODE)). Text and diagrams are CC-BY-SA-4.0
([LICENSE-DOCS](LICENSE-DOCS)).
