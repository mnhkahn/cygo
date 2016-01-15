package container

import (
	"sync"
)

// 定义默认大小的chunk
const chunkSize int = 10240

type chunk struct {
	items [chunkSize]interface{}
	first int
	last  int
	next  *chunk
}

type Queue struct {
	head  *chunk
	tail  *chunk
	count int
	sync.Mutex
}

func NewQueue() *Queue {

	ck := new(chunk)
	queue := &Queue{
		head:  ck,
		tail:  ck,
		count: 0,
	}
	return queue
}

// 向队列中添加对象
func (q *Queue) Push(item interface{}) {

	q.Lock()
	defer q.Unlock()

	// 队列不接受空对象
	if nil == item {
		return
	}

	// chunk已经满，需要new一个新的chunk
	if q.tail.last >= chunkSize {
		q.tail.next = new(chunk)
		q.tail = q.tail.next
	}

	q.tail.items[q.tail.last] = item
	q.tail.last++
	q.count++

}

func (q *Queue) Pop() interface{} {

	q.Lock()
	defer q.Unlock()

	// 队列为空啊
	if q.count == 0 {
		return nil
	}

	item := q.head.items[q.head.first]
	q.head.first++
	q.count--

	// 某个trunk块取空
	if q.head.first >= q.head.last {

		if q.count == 0 {
			q.head.first = 0
			q.head.last = 0
			q.head.next = nil
		} else {
			// 移动到下一个chunk块
			q.head = q.head.next
		}

	}

	return item

}

func (q *Queue) Len() int {

	q.Lock()
	defer q.Unlock()
	ct := q.count
	return ct
}
