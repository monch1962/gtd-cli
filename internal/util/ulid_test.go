package util

import (
	"strings"
	"testing"
	"time"
)

func TestRealIDGenerator(t *testing.T) {
	clock := NewFixedClock(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))
	gen := NewRealIDGenerator(clock)

	id := gen.NewID("tsk_")

	if !strings.HasPrefix(id, "tsk_") {
		t.Errorf("ID should have prefix tsk_, got %s", id)
	}

	ulidPart := strings.TrimPrefix(id, "tsk_")
	if len(ulidPart) != 26 {
		t.Errorf("ULID part should be 26 chars, got %d", len(ulidPart))
	}
}

func TestDeterministicIDGenerator(t *testing.T) {
	gen := NewDeterministicIDGenerator()

	id1 := gen.NewID("tsk_")
	id2 := gen.NewID("prj_")

	if !strings.HasPrefix(id1, "tsk_") {
		t.Errorf("ID1 should have prefix tsk_, got %s", id1)
	}
	if !strings.HasPrefix(id2, "prj_") {
		t.Errorf("ID2 should have prefix prj_, got %s", id2)
	}

	if id1 == id2 {
		t.Error("IDs should be different")
	}
}

func TestDeterministicIDGenerator_Sequence(t *testing.T) {
	gen := NewDeterministicIDGenerator()

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := gen.NewID("tsk_")
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}
