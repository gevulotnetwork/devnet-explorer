package stats

import (
	"math/rand"
	"testing"
	"time"
)

func TestMonotonicDay(t *testing.T) {
	now := time.Now()
	prev := now.AddDate(0, 0, -1)

	for i := 0; i < 1000; i++ {
		if monotonicDay(now) <= monotonicDay(prev) {
			t.Fatalf("monotonicDay(now) %q <= monotonicDay(prev) %q", now.Format(time.DateTime), prev.Format(time.DateTime))
		}

		// Update prev with current time.
		prev = now

		// Add one day to current day.
		now = now.AddDate(0, 0, 1)

		// Add random number of minutes to day, to cause random time of day shifting.
		now = now.Add(time.Minute * time.Duration(rand.Int63n(240)))
	}
}
