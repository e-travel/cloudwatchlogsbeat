package beater

import (
	"regexp"
	"time"

	"github.com/elastic/beats/libbeat/logp"
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

// Validates a multiline configuration section
func ValidateMultiline(multiline *Multiline) {
	if multiline == nil {
		return
	}

	// Check if valid regular expression for multiline
	_, err := regexp.Compile(multiline.Pattern)
	Fatal(err)

	// Check match mode
	match := multiline.Match
	switch match {
	case "after":
	case "before":
	default:
		panic("Configuration: Invalid match type in multiline mode: " + match)
	}
}

// returns true if the event timestamp is before the specified horizon
func IsBefore(horizon time.Duration, lastEventTimestamp int64) bool {
	horizonTimestamp := time.Now().UTC().Add(-horizon)
	return ToTime(lastEventTimestamp).Before(horizonTimestamp)
}
