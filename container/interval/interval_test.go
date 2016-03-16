package interval

import "testing"

func TestInterval(t *testing.T) {
	interval := NewInterval()

	// interval.Add(New(0, 1))
	// interval.DebugPrint()
	// println("exists", interval.Exists(New(2, 2)))

	// interval.Add(New(1, 2))
	// interval.DebugPrint()
	// println("exists", interval.Exists(New(2, 2)))

	interval.Add(New(307200, 921599))
	interval.Add(New(1024000, 2662399))
	interval.DebugPrint()

	// interval.Sub(New(2, 3))
	interval.Sub(New(307200, 409599))
	interval.DebugPrint()
	// println("exists", interval.Exists(New(2, 2)))
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
