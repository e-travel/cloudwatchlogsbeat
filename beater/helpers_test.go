package beater

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_IsBefore_ReturnsFalse_WhenTimestamp_IsWithinTheHorizon(t *testing.T) {
	horizon := time.Hour
	lastEventTimestamp := TimeBeforeNowInMilliseconds(30 * time.Minute)
	assert.False(t, IsBefore(horizon, lastEventTimestamp))
}

func Test_IsBefore_ReturnsFalse_WhenTimestamp_IsNotWithinTheHorizon(t *testing.T) {
	horizon := time.Hour
	lastEventTimestamp := TimeBeforeNowInMilliseconds(90 * time.Minute)
	assert.True(t, IsBefore(horizon, lastEventTimestamp))
}
