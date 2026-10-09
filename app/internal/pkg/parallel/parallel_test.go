package parallel

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMap_AllSucceed(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8}

	got, errs := Map(context.Background(), in, len(in), func(x int) (int, error) {
		return x * 10, nil
	})

	assert.Equal(t, []int{10, 20, 30, 40, 50, 60, 70, 80}, got)
	assert.Empty(t, errs)
}

func TestMap_EmptyInput(t *testing.T) {
	var calls atomic.Int32

	got, errs := Map(context.Background(), []int{}, 1, func(x int) (int, error) {
		calls.Add(1)
		return x, nil
	})

	assert.Empty(t, got)
	assert.Empty(t, errs)
	assert.Zero(t, calls.Load())
}

func TestMap_CallsFnExactlyOncePerElement(t *testing.T) {
	in := make([]int, 50)
	for i := range in {
		in[i] = i
	}

	counts := make([]atomic.Int32, len(in))

	_, _ = Map(context.Background(), in, 7, func(x int) (int, error) {
		counts[x].Add(1)
		return x, nil
	})

	for i := range counts {
		assert.Equal(t, int32(1), counts[i].Load(), "element %d", i)
	}
}

func TestMap_CollectsErrors(t *testing.T) {
	errA := errors.New("a")
	errB := errors.New("b")

	cases := []struct {
		name string

		failures map[int]error

		wantErrs []error
	}{
		{name: "no failures", failures: nil, wantErrs: nil},
		{name: "one failure", failures: map[int]error{2: errA}, wantErrs: []error{errA}},
		{name: "two failures", failures: map[int]error{0: errA, 3: errB}, wantErrs: []error{errA, errB}},
		{name: "every element fails", failures: map[int]error{0: errA, 1: errA, 2: errA, 3: errA}, wantErrs: []error{errA, errA, errA, errA}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := []int{0, 1, 2, 3}

			_, errs := Map(context.Background(), in, len(in), func(x int) (int, error) {
				return x, tc.failures[x]
			})

			require.Len(t, errs, len(tc.wantErrs))
			for _, want := range tc.wantErrs {
				assert.True(t, containsIs(errs, want), "missing %v", want)
			}
		})
	}
}

func TestMap_PreservesOrderWhenCompletionIsReversed(t *testing.T) {
	const n = 5
	in := []int{0, 1, 2, 3, 4}

	// Element i finishes only after element i+1 has finished.
	done := make([]chan struct{}, n+1)
	for i := range done {
		done[i] = make(chan struct{})
	}
	close(done[n])

	got, errs := Map(context.Background(), in, n, func(x int) (string, error) {
		<-done[x+1]
		close(done[x])
		return fmt.Sprintf("r%d", x), nil
	})

	assert.Equal(t, []string{"r0", "r1", "r2", "r3", "r4"}, got)
	assert.Empty(t, errs)
}

func TestMap_AllCallsRunConcurrentlyWhenLimitEqualsLen(t *testing.T) {
	const n = 6
	in := make([]int, n)

	var arrived atomic.Int32
	barrier := make(chan struct{})

	finished := make(chan struct{})
	go func() {
		defer close(finished)
		_, _ = Map(context.Background(), in, n, func(x int) (int, error) {
			if arrived.Add(1) == n {
				close(barrier)
			}
			<-barrier
			return x, nil
		})
	}()

	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock: not all calls ran concurrently")
	}
}

func TestMap_RespectsConcurrencyLimit(t *testing.T) {
	cases := []struct {
		name string

		items int
		limit int
	}{
		{name: "limit 1 is sequential", items: 5, limit: 1},
		{name: "limit 2", items: 8, limit: 2},
		{name: "limit 3 with 10 items", items: 10, limit: 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var current, peak atomic.Int32
			started := make(chan struct{}, tc.items)
			release := make(chan struct{})

			in := make([]int, tc.items)

			finished := make(chan struct{})
			go func() {
				defer close(finished)
				_, _ = Map(context.Background(), in, tc.limit, func(x int) (int, error) {
					c := current.Add(1)
					for {
						p := peak.Load()
						if c <= p || peak.CompareAndSwap(p, c) {
							break
						}
					}
					started <- struct{}{}
					<-release
					current.Add(-1)
					return x, nil
				})
			}()

			// Wait until the limit is saturated; an over-run is caught by the peak check below.
			for i := 0; i < tc.limit; i++ {
				select {
				case <-started:
				case <-time.After(5 * time.Second):
					t.Fatal("timed out waiting for workers to start")
				}
			}

			go func() {
				// Keep releasing workers until Map returns.
				for {
					select {
					case release <- struct{}{}:
					case <-finished:
						return
					}
				}
			}()

			select {
			case <-finished:
			case <-time.After(5 * time.Second):
				t.Fatal("deadlock waiting for Map to finish")
			}

			assert.LessOrEqual(t, peak.Load(), int32(tc.limit))
			assert.Equal(t, int32(0), current.Load())
		})
	}
}

func containsIs(errs []error, target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
	}

	return false
}
