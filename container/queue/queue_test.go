package queue

import "testing"

func TestQueue(t *testing.T) {
	len := 20
	queue := NewQueue()
	for i := 0; i < len; i++ {
		queue.Push(i)
	}

	if queue.Len() != len {
		t.Error("Len:", queue.Len())
	}

	queue.Debug()

	queue.Clear()

	queue.Debug()
}
