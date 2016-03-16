// 区间实现
package interval

import "fmt"

type IntervalBlockIface interface {
	Start() int64
	End() int64
	SetStart(start int64)
	SetEnd(end int64)
}

type Interval struct {
	intervals []IntervalBlockIface
}

func NewInterval() *Interval {
	interval := new(Interval)
	return interval
}

func (this *Interval) DebugPrint() {
	for _, interval := range this.intervals {
		fmt.Printf("[%d~%d] ", interval.Start(), interval.End())
	}
	fmt.Println()
}

func (this *Interval) Len() int {
	return len(this.intervals)
}

func (this *Interval) Empty() bool {
	return len(this.intervals) == 0
}

func (this *Interval) Get() []IntervalBlockIface {
	return this.intervals
}

func (this *Interval) Exists(intervalB IntervalBlockIface) bool {
	for _, interval := range this.intervals {
		if Include(interval, intervalB) {
			return true
		}
	}
	return false
}

func (this *Interval) In(b int64) bool {
	for _, interval := range this.intervals {
		if In(b, interval) {
			return true
		}
	}
	return false
}

// 插入排序
func (this *Interval) Add(intervalB IntervalBlockIface) {
	insertIndex := len(this.intervals)
	for i, interval := range this.intervals {
		// fmt.Println(Include(interval, intervalB), Include(intervalB, interval), interval.End() == intervalB.Start(), interval.Start() == intervalB.End(), intervalB.Start() < interval.Start(), "AAAAAAAAAAAAA")
		if Include(interval, intervalB) { // a包含b
			return
		} else if Include(intervalB, interval) { // b包含a
			this.intervals[i] = intervalB
			return
		} else if interval.End() == intervalB.Start() || interval.End()+1 == intervalB.Start() { // merge a b
			interval.SetEnd(intervalB.End())
			return
		} else if interval.Start() == intervalB.End() || interval.Start() == intervalB.End()+1 { // merge b a
			interval.SetStart(intervalB.Start())
			return
		} else if intervalB.Start() < interval.Start() { // insert
			insertIndex = i
			break
		}
	}
	if insertIndex >= 0 {
		// https://github.com/golang/go/wiki/SliceTricks
		this.intervals = append(this.intervals[:insertIndex], append([]IntervalBlockIface{intervalB}, this.intervals[insertIndex:]...)...)
	}
}

func (this *Interval) Sub(intervalB IntervalBlockIface) {
	for i, interval := range this.intervals {
		// b include a
		if Include(intervalB, interval) {
			// println(i, "AAAAAAAAAAAAAAAAAAAAA")
			this.intervals = append(this.intervals[:i], this.intervals[i+1:]...)
		} else {
			Sub(interval, intervalB)
		}
	}
}

// Common function
// 是否包含
func Include(a, b IntervalBlockIface) bool {
	return a.Start() <= b.Start() && a.End() >= b.End()
}

func In(a int64, b IntervalBlockIface) bool {
	return a >= b.Start() && a <= b.End()
}

func Equal(a, b IntervalBlockIface) bool {
	return a.Start() == b.Start() && a.End() == b.End()
}

// a - b 的绝对值
func Sub(a, b IntervalBlockIface) IntervalBlockIface {
	if a.Start() <= b.Start() { // 减前面
		a.SetStart(b.End() + 1)
	} else if a.End() >= b.End() { // 减后面
		a.SetEnd(b.Start() - 1)
	}
	return a
}
