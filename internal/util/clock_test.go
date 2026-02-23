package util

import (
	"testing"
	"time"
)

func TestFixedClock(t *testing.T) {
	fixed := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := NewFixedClock(fixed)

	if got := clock.Now(); !got.Equal(fixed) {
		t.Errorf("FixedClock.Now() = %v, want %v", got, fixed)
	}
}

func TestRealClock(t *testing.T) {
	clock := RealClock{}
	before := time.Now()
	got := clock.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Errorf("RealClock.Now() = %v, expected between %v and %v", got, before, after)
	}
}
