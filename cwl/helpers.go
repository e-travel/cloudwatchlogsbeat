package cwl

import (
	"time"

	"github.com/elastic/beats/v7/libbeat/logp"
)

// Checks for error existance and panics
func Fatal(err error) {
	if err != nil {
		logp.Critical(err.Error())
		panic(err)
	}
}

// Converts a timestamp to time
func ToTime(timestamp int64) time.Time {
	return time.Unix(timestamp/1000, (timestamp%1000)*1000000)
}

// returns true if the event timestamp is before the specified horizon
func IsBefore(horizon time.Duration, lastEventTimestamp int64) bool {
	horizonTimestamp := time.Now().UTC().Add(-horizon)
	return ToTime(lastEventTimestamp).Before(horizonTimestamp)
}
