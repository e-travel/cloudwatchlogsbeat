package beater

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_IsBefore_ReturnsFalse_WhenTimestamp_IsAfterTheHorizon(t *testing.T) {
	horizon := time.Hour
	lastEventTimestamp := TimeBeforeNowInMilliseconds(30 * time.Minute)
	assert.False(t, IsBefore(horizon, lastEventTimestamp))
}

func Test_IsBefore_ReturnsTrue_WhenTimestamp_IsBeforeTheHorizon(t *testing.T) {
	horizon := time.Hour
	lastEventTimestamp := TimeBeforeNowInMilliseconds(90 * time.Minute)
	assert.True(t, IsBefore(horizon, lastEventTimestamp))
}

func Test_IsBefore_ReturnsTrue_WhenHorizon_IsZero_And_Timestamp_IsNow(t *testing.T) {
	horizon := time.Duration(0)
	lastEventTimestamp := TimeBeforeNowInMilliseconds(0 * time.Minute)
	assert.True(t, IsBefore(horizon, lastEventTimestamp))
}
