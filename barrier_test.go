package barrier_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/db7/barrier"
	"github.com/facebookgo/ensure"
)

func TestBarrier_manyrounds(t *testing.T) {
	var count int64
	rounds := 100
	n := 10 // number of goroutines
	b := barrier.New(n)

	for i := 0; i < rounds; i++ {
		var wg sync.WaitGroup
		wg.Add(n)
		for j := 0; j < n; j++ {
			go func() {
				b.Await(func() error {
					atomic.AddInt64(&count, 1)
					return nil
				})
				wg.Done()
			}()
		}
		wg.Wait()
	}
	ensure.True(t, atomic.LoadInt64(&count) == int64(rounds))
}

func TestBarrier_abortBeforeLast(t *testing.T) {
	n := 10 // number of goroutines
	b := barrier.New(n)

	// one round before abort
	var wg sync.WaitGroup
	wg.Add(n)
	for j := 0; j < n; j++ {
		go func() {
			defer wg.Done()
			err := b.Await(nil)
			ensure.Nil(t, err)
		}()
	}
	wg.Wait()

	// one round when abort is called
	wg.Add(n)
	for j := 0; j < n-1; j++ {
		go func() {
			defer wg.Done()
			err := b.Await(nil)
			ensure.True(t, err == barrier.ErrBarrierAborted)
		}()
	}
	// last goroutine
	b.Abort()
	go func() {
		defer wg.Done()
		err := b.Await(nil)
		ensure.True(t, err == barrier.ErrBarrierAborted)
	}()
	wg.Wait()

	// one last round where all goroutines should fail (aborted)
	wg.Add(n)
	for j := 0; j < n; j++ {
		go func() {
			defer wg.Done()
			err := b.Await(nil)
			ensure.True(t, err == barrier.ErrBarrierAborted)
		}()
	}
	wg.Wait()
}

func ExampleBarrier_simple() {
	n := 40 // number of goroutines
	b := barrier.New(n)

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(k int) {
			b.Await(func() error {
				fmt.Println(k, "is the last goroutine")
				return nil
			})
			wg.Done()
		}(i)
	}
	wg.Wait()
}
