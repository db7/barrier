// Package barrier provides a data structure to synchronize a group of
// goroutines, blocking them until all of the arrive on the barrier. Once the
// last goroutine arrive an optional callback is executed in isolation.
package barrier

import (
	"fmt"
	"sync"
)

// ErrBarrierAborted is returned by Await() if Abort() was called.
var ErrBarrierAborted = fmt.Errorf("Barrier aborted")

// ErrBarrierMisused is returned by Await() if more than n concurrent Await()
// calls are detected.
var ErrBarrierMisused = fmt.Errorf("Barrier misused: more than n concurrent Await() calls")

// Callback is called by the last goroutine entering the barrier.
type Callback func() error

// Barrier synchronizes a group of goroutines and optinally executes a callback
// in isolation.
type Barrier struct {
	sync.Mutex
	n     int64
	count int64
	done  chan bool
	abort chan bool
}

// New returns a new Barrier which expects n goroutines to synchronize.
func New(n int) *Barrier {
	return &Barrier{
		n:     int64(n),
		count: int64(n),
		done:  make(chan bool),
		abort: make(chan bool),
	}
}

// Abort marks the barrier as aborted and signal all waiting goroutines.
// The barrier cannot be reset once aborted.
func (b *Barrier) Abort() {
	close(b.abort)
}

// Await synchronizes n goroutines and executes in isolation the callback of
// the last goroutine calling Await. Await returns any error the callback
// returns to one goroutine; if Abort() is called, ErrBarrierAborted is
// returned. The number of goroutines call Await should always match the value
// n passed in the barrier's initialization.
func (b *Barrier) Await(cb Callback) error {
	if b.aborted() {
		return ErrBarrierAborted
	}
	// keep copy of current state
	b.Lock()
	b.count--
	count := b.count
	done := b.done
	b.Unlock()

	// more than n goroutines called Await
	if count < 0 {
		b.Abort()
		return ErrBarrierMisused
	}

	// wait for others and for callback execution to finish
	if count > 0 {
		return b.wait(done)
	}

	// execute callback if last goroutine
	var err error
	if cb != nil {
		err = cb()
	}
	b.reset()
	return err
}

// aborted checks whether Barrier is aborted
func (b *Barrier) aborted() bool {
	select {
	case <-b.abort:
		return true
	default:
		return false
	}
}

// wait waits for execution of callback or abort()
func (b *Barrier) wait(done chan bool) error {
	select {
	case <-done:
		if b.aborted() {
			// guarantee that all blocking goroutines return ErrBarrierAborted if
			// barrier was aborted
			return ErrBarrierAborted
		}
		return nil
	case <-b.abort:
		return ErrBarrierAborted
	}
}

// reset resets the barrier for another round
func (b *Barrier) reset() {
	b.Lock()
	close(b.done)
	b.done = make(chan bool)
	b.count = b.n
	b.Unlock()
}
