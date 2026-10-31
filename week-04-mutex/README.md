# Week 4: Mutexes and the runtime semaphore

What happens when a goroutine cannot acquire a mutex: the CAS fast path, spinning, parking
on the runtime semaphore, starvation mode and handoff. Also `RWMutex`, lock ordering,
deadlock and livelock, and how these mechanisms inform concurrent map design.

The runnable demos are in [demos/](demos/). The practice session, source prediction,
interview questions and map design, is in [practice/](practice/). There is no new graded
homework this week. The slides come later.
