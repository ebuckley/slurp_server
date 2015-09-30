package LRU

import (
	"container/heap"
	"fmt"
	"time"
)

type LRUHeap []*Record

func (h LRUHeap) Len() int {
	return len(h)
}
func (h LRUHeap) Less(i, j int) bool {
	return h[i].accessed.Before(h[j].accessed)
}

func (h LRUHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]

	h[i].index = i
	h[j].index = j
}

func (h *LRUHeap) Push(item interface{}) {
	//TODO count the total size hear, pop from the heap if the new size is too large
	rec := item.(*Record)
	rec.index = len(*h)
	*h = append(*h, rec)
}
func (h *LRUHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	item.index = -1
	*h = old[0 : n-1]
	return item
}

func (h *LRUHeap) update(item Record) {
	heap.Fix(h, item.index)
}

type Record struct {
	name     string
	data     []byte
	accessed time.Time //priority
	index    int       //heap index
}

func (r Record) String() string {
	return fmt.Sprintf("accessed: %s index: %s name: %s", r.accessed, r.index, r.name)
}

type Cache struct {
	heap        *LRUHeap
	cacheSet    map[string]*Record
	maxSize     int
	currentSize int
}

func NewCache(size int) (lru *Cache) {
	lru = new(Cache)
	lru.heap = &LRUHeap{}
	lru.maxSize = size
	lru.currentSize = 0
	lru.cacheSet = make(map[string]*Record)

	heap.Init(lru.heap)
	return lru
}

func (c *Cache) Push(name string, data []byte) {
	if c.cacheSet[name] != nil {
		// TODO handle error in a better way, decide if this should be handled at all?
		return
	}

	//pop LRU from the heap if we will be exceeding the Cache size
	if c.currentSize+len(data) > c.maxSize {
		oldRec := heap.Pop(c.heap).(*Record)
		c.currentSize = c.currentSize - len(oldRec.data)
		delete(c.cacheSet, name)
	}

	rec := new(Record)
	rec.accessed = time.Now()
	rec.data = data
	rec.name = name
	heap.Push(c.heap, rec)

	c.currentSize = c.currentSize + len(rec.data)
	c.cacheSet[name] = rec
}

func (c *Cache) Get(name string) ([]byte, bool) {

	if rec := c.cacheSet[name]; rec != nil {
		rec.accessed = time.Now()
		c.heap.update(*rec)
		return rec.data, true
	}
	return nil, false
}
