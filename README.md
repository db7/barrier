# barrier

[![GoDoc](https://godoc.org/github.com/db7/barrier?status.png)](https://godoc.org/github.com/db7/barrier)

Barrier is a data structure to synchronize a fixed-size group of goroutines,
blocking them until all arrive on the barrier, by calling `Await()`.
Once the last goroutine arrives, an optional callback is executed in isolation.
The barrier can be immediately reused after the goroutines return from
`Await()`.

This barrier implementation supports early termination by calling `Abort()`.
Once aborted the barrier cannot be reused. Also, if the barrier detects that
`Await()` is concurrently called by often than the number specified in the
constructor, the barrier is aborted.

### Algorithm

The algorithm is pretty similar to the simple algorithm described in
*The Art of Multiprocessor Programming - Herlihy and Shavit*.
By using a mutex and creating a new channel on every reset, however,
one does not have to implement a goroutine-bound sense variable.

### Alternatives

A simpler version, also using channels and a mutex, can be found
[here](http://www.gofragments.net/client/blog/concurrency/2015/11/09/cyclicBarrierMutexChan/index.html).
