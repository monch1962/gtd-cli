package util

import (
	"crypto/rand"
	"io"
	"sync/atomic"
	"time"

	"github.com/oklog/ulid/v2"
)

type IDGenerator interface {
	NewID(prefix string) string
}

type RealIDGenerator struct {
	entropy io.Reader
	clock   Clock
}

func NewRealIDGenerator(clk Clock) *RealIDGenerator {
	if clk == nil {
		clk = RealClock{}
	}
	return &RealIDGenerator{
		entropy: ulid.Monotonic(rand.Reader, 0),
		clock:   clk,
	}
}

func (g *RealIDGenerator) NewID(prefix string) string {
	id := ulid.MustNew(ulid.Timestamp(g.clock.Now()), g.entropy)
	return prefix + id.String()
}

type DeterministicIDGenerator struct {
	counter uint64
	base    string
}

func NewDeterministicIDGenerator() *DeterministicIDGenerator {
	return &DeterministicIDGenerator{
		counter: 0,
		base:    "01HXYZ00000000000000000000",
	}
}

func (g *DeterministicIDGenerator) NewID(prefix string) string {
	n := atomic.AddUint64(&g.counter, 1)
	ts := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	id := ulid.MustNew(ulid.Timestamp(ts), seqReader(n))
	return prefix + id.String()
}

type seqReader uint64

func (s seqReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(uint64(s) + uint64(i))
	}
	return len(p), nil
}
