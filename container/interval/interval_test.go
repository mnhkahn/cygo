package interval

import "testing"

func TestInterval(t *testing.T) {
	interval := NewInterval()

	// interval.Add(New(307200, 511999))
	// interval.DebugPrint()

	// interval.Sub(New(409600, 511999))
	// interval.DebugPrint()

	interval.Add(New(0, 1397587))
	interval.DebugPrint()

	interval.Sub(New(0, 102399))
	interval.DebugPrint()
}

type TestIntervalStr struct {
	start int64
	end   int64
}

func New(s, e int64) *TestIntervalStr {
	n := new(TestIntervalStr)
	n.start, n.end = s, e
	return n
}

func (this *TestIntervalStr) Start() int64 {
	return this.start
}

func (this *TestIntervalStr) End() int64 {
	return this.end
}

func (this *TestIntervalStr) SetStart(start int64) {
	this.start = start
}

func (this *TestIntervalStr) SetEnd(end int64) {
	this.end = end
}
