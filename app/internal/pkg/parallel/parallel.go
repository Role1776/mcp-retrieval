package parallel

import (
	"context"
	"sync"
)

func Map[In, Out any](ctx context.Context, in []In, limit int, fn func(In) (Out, error)) ([]Out, []error) {
	results := make([]Out, len(in))
	errs := make([]error, 0, len(in))

	if limit <= 0 {
		limit = len(in)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, limit)

	stopped := len(in)

loop:
	for i, item := range in {
		select {
		case <-ctx.Done():
			stopped = i

			break loop
		case sem <- struct{}{}:
		}

		wg.Go(func() {
			defer func() {
				<-sem
			}()

			result, err := fn(item)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
			results[i] = result
		})
	}
	wg.Wait()

	for range in[stopped:] {
		errs = append(errs, ctx.Err())
	}

	return results, errs
}
