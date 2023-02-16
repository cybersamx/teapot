package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime_DurationSeconds(t *testing.T) {
	tests := []struct {
		timestamp time.Duration
		want      int
	}{
		{time.Nanosecond, 0},
		{time.Microsecond, 0},
		{time.Millisecond, 0},
		{time.Second, 1},
		{time.Minute, 60},
		{time.Hour, 3600},
	}

	for _, tc := range tests {
		seconds := ToSeconds(tc.timestamp)
		assert.Equal(t, tc.want, seconds)
	}
}
