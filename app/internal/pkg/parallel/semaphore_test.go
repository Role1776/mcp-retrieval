package parallel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemaphore_ZeroLimitNeverBlocks(t *testing.T) {
	sem := NewSemaphore(0)

	for range 100 {
		_, err := sem.Acquire(context.Background())
		require.NoError(t, err)
	}
}

func TestSemaphore_BlocksAtLimitUntilRelease(t *testing.T) {
	sem := NewSemaphore(1)

	release, err := sem.Acquire(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = sem.Acquire(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	release()

	_, err = sem.Acquire(context.Background())
	assert.NoError(t, err)
}
