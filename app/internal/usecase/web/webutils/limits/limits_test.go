package limits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveMax(t *testing.T) {
	cases := []struct {
		name string

		requested    int
		defaultValue int
		maximum      int

		want int
	}{
		{
			name:      "requested within bounds is kept",
			requested: 3, defaultValue: 5, maximum: 20,
			want: 3,
		},
		{
			name:      "requested above maximum with default 0",
			requested: 43, defaultValue: 0, maximum: 2,
			want: 2,
		},
		{
			name:      "negative requested above maximum with default 0",
			requested: -43, defaultValue: 0, maximum: 2,
			want: 0,
		},
		{
			name:      "zero falls back to the default",
			requested: 0, defaultValue: 5, maximum: 20,
			want: 5,
		},
		{
			name:      "negative falls back to the default",
			requested: -1, defaultValue: 5, maximum: 20,
			want: 5,
		},
		{
			name:      "requested above the maximum is capped",
			requested: 999, defaultValue: 5, maximum: 20,
			want: 20,
		},
		{

			name:      "default above the maximum is capped",
			requested: 0, defaultValue: 5, maximum: 1,
			want: 1,
		},
		{
			name:      "requested above a maximum below the default is capped",
			requested: 3, defaultValue: 5, maximum: 1,
			want: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveMax(tc.requested, tc.defaultValue, tc.maximum)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestResolveMinMax(t *testing.T) {
	cases := []struct {
		name string

		requested    int64
		defaultValue int64
		minimum      int64
		maximum      int64

		want int64
	}{
		{
			name:      "requested within bounds is kept",
			requested: 5000, defaultValue: 5000, minimum: 1000, maximum: 10000,
			want: 5000,
		},
		{
			name:      "zero falls back to the default",
			requested: 0, defaultValue: 5000, minimum: 1000, maximum: 10000,
			want: 5000,
		},
		{
			name:      "negative requested falls back to the default",
			requested: -500, defaultValue: 5000, minimum: 1000, maximum: 10000,
			want: 5000,
		},
		{
			name:      "requested exactly equal to minimum is kept",
			requested: 1000, defaultValue: 5000, minimum: 1000, maximum: 10000,
			want: 1000,
		},
		{
			name:      "requested exactly equal to maximum is kept",
			requested: 10000, defaultValue: 5000, minimum: 1000, maximum: 10000,
			want: 10000,
		},
		{
			name:      "default above the maximum is capped",
			requested: 0, defaultValue: 99000, minimum: 1000, maximum: 10000,
			want: 10000,
		},

		{
			name:      "requested below the minimum is raised",
			requested: 10, defaultValue: 5000, minimum: 1000, maximum: 10000,
			want: 1000,
		},
		{
			name:      "requested above the maximum is capped",
			requested: 99000, defaultValue: 5000, minimum: 1000, maximum: 10000,
			want: 10000,
		},
		{
			name:      "default below the minimum is raised",
			requested: 0, defaultValue: 10, minimum: 1000, maximum: 10000,
			want: 1000,
		},
		{

			name:      "minimum above the maximum yields the maximum",
			requested: 5000, defaultValue: 5000, minimum: 20000, maximum: 10000,
			want: 10000,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveMinMax(tc.requested, tc.defaultValue, tc.minimum, tc.maximum)

			assert.Equal(t, tc.want, got)
		})
	}
}
