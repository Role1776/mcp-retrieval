package parallel

import "sync"

func Map[In, Out any](in []In, fn func(In) (Out, error)) ([]Out, []error) {
	results := make([]Out, len(in))
	errs := make([]error, 0, len(in))

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, item := range in {
		wg.Go(func() {
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

	return results, errs
}
