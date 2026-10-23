package parallel

import "context"

type Semaphore struct {
	slots chan struct{}
}

func NewSemaphore(limit int) *Semaphore {
	if limit <= 0 {
		return &Semaphore{}
	}

	return &Semaphore{slots: make(chan struct{}, limit)}
}

func (s *Semaphore) Acquire(ctx context.Context) (release func(), err error) {
	if s.slots == nil {
		return func() {}, nil
	}

	select {
	case s.slots <- struct{}{}:
		return func() { <-s.slots }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
