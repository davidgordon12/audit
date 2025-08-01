package audit

import (
	"errors"
)

const capacity = 3

// Only provides the following functionalitiy.
// Append, Pop
type Queue struct {
	buff  []string
	head  int
	tail  int
	count int
}

func NewQueue() *Queue {
	q := new(Queue)
	q.buff = make([]string, capacity)
	q.head = 0
	q.tail = 0
	q.count = 0
	return q
}

// Appends an item to the buffer.
// If the buffer is full, it overwrites the oldest element.
func (q *Queue) Append(s string) {
	q.buff[q.tail] = s

	if q.count == capacity {
		q.head = (q.head + 1) % capacity
	}

	q.tail = (q.tail + 1) % capacity

	if q.count < capacity {
		q.count++
	}
}

// Removes an item from the beginning of the queue and returns the item.
func (q *Queue) Pop() (string, error) {
	if q.count == 0 {
		return "", errors.New("cannot pop from empty queue")
	}

	s := q.buff[q.head]
	q.head = (q.head + 1) % capacity
	q.count--

	return s, nil
}
