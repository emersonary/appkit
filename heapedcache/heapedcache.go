package heapedcache

import (
	"container/heap"
	"sync"
	"time"
)

type Item[TId any, TObj any] struct {
	ID        TId
	index     int
	Refreshed time.Time
	obj       *TObj
}

type items[TId any, TObj any] []*Item[TId, TObj]

type HeapedCache[TId any, TObj any] struct {
	mu         sync.RWMutex
	maxRows    int
	mapItems   map[any]*Item[TId, TObj]
	sliceItems items[TId, TObj]
}

// New creates a fixed-size, thread-safe LRU-style cache backed by a min-heap.
func New[TId any, TObj any](maxRows int) *HeapedCache[TId, TObj] {
	return &HeapedCache[TId, TObj]{
		maxRows:    maxRows,
		mapItems:   make(map[any]*Item[TId, TObj], maxRows+1),
		sliceItems: make(items[TId, TObj], 0, maxRows+1),
	}
}

// NewHeapedCache is an alias for New, matching the original heapedcache API.
func NewHeapedCache[TId any, TObj any](maxRows int) *HeapedCache[TId, TObj] {
	return New[TId, TObj](maxRows)
}

func (t *HeapedCache[TId, TObj]) pop() *TObj {
	item := heap.Pop(&t.sliceItems).(*Item[TId, TObj])
	delete(t.mapItems, item.ID)
	return item.obj
}

func (t *HeapedCache[TId, TObj]) Pop() *TObj {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.pop()
}

func (t *HeapedCache[TId, TObj]) PopWithRefreshed() (*TObj, time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.popWithRefreshed()
}

func (t *HeapedCache[TId, TObj]) popWithRefreshed() (*TObj, time.Time) {
	item := heap.Pop(&t.sliceItems).(*Item[TId, TObj])
	delete(t.mapItems, item.ID)
	return item.obj, item.Refreshed
}

func (t *HeapedCache[TId, TObj]) Get(id any) *TObj {
	t.mu.Lock()
	defer t.mu.Unlock()

	item := t.mapItems[id]
	if item == nil {
		return nil
	}

	return item.obj
}

func (t *HeapedCache[TId, TObj]) GetOrAdd(id TId, fn func(id TId) *TObj) *TObj {
	t.mu.Lock()
	defer t.mu.Unlock()

	findItem := t.mapItems[id]
	if findItem == nil {
		result := fn(id)
		if result == nil {
			return nil
		}

		return t.push(id, result)
	}

	return findItem.obj
}

func (t *HeapedCache[TId, TObj]) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return len(t.mapItems)
}

func (t *HeapedCache[TId, TObj]) push(id TId, item *TObj) *TObj {
	if item == nil {
		return nil
	}

	findItem := t.mapItems[id]
	if findItem == nil {
		newItem := &Item[TId, TObj]{
			ID:        id,
			index:     len(t.sliceItems),
			Refreshed: time.Now(),
			obj:       item,
		}

		t.mapItems[id] = newItem
		heap.Push(&t.sliceItems, newItem)

		if len(t.sliceItems) > t.maxRows {
			t.pop()
		}
	} else {
		findItem.obj = item
		findItem.Refreshed = time.Now()
		heap.Fix(&t.sliceItems, findItem.index)
	}

	return item
}

func (t *HeapedCache[TId, TObj]) Push(id TId, item *TObj) *TObj {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.push(id, item)
}

func (t *HeapedCache[TId, TObj]) Remove(id TId) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	findItem := t.mapItems[id]
	if findItem == nil {
		return false
	}

	heap.Remove(&t.sliceItems, findItem.index)
	delete(t.mapItems, id)
	return true
}

func (h *items[TId, TObj]) Len() int {
	return len(*h)
}

func (h *items[TId, TObj]) Less(i, j int) bool {
	return (*h)[i].Refreshed.Compare((*h)[j].Refreshed) < 0
}

func (h *items[TId, TObj]) Swap(i, j int) {
	if i == j {
		return
	}

	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
	(*h)[i].index = i
	(*h)[j].index = j
}

func (h *items[TId, TObj]) Push(x any) {
	*h = append(*h, x.(*Item[TId, TObj]))
}

func (h *items[TId, TObj]) Pop() any {
	n := len(*h)
	item := (*h)[n-1]
	(*h)[n-1] = nil
	item.index = -1
	*h = (*h)[0 : n-1]

	return item
}
