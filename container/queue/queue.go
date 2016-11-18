/*
 * @Author: lichao115
 * @Date: 2016-11-18 11:29:55
 * @Last Modified by: lichao115
 * @Last Modified time: 2016-11-18 11:36:46
 */

package queue

import (
	"log"
	"sync"
)

// 定义默认大小的chunk
const chunkSize int = 3

// const chunkSize int = 1024

type chunk struct {
	items [chunkSize]interface{}
	first int // 头指针
	last  int // 尾指针
	next  *chunk
}

// 这是一个二级队列，队列里面是桶，每个桶有chunkSize个元素
type Queue struct {
	head  *chunk // 头指针
	tail  *chunk // 尾指针
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

func (q *Queue) Clear() {
	for item := q.Pop(); item != nil; item = q.Pop() {
	}
}

func (q *Queue) Debug() {
	// 队列为空啊
	if q.count == 0 {
		return
	}

	i := q.head.first
	head := q.head
	for c := 0; c < q.Len(); c++ {
		if i >= head.last {
			head = head.next
			i = head.first
		}

		item := head.items[i]
		i++

		log.Println(c, item)
	}
}
