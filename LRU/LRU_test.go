package LRU

import (
	"testing"
)

func compare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i, val := range a {
		if val != b[i] {
			return false
		}
	}
	return true
}

func TestCreateNewLRU(t *testing.T) {
	var size int = 1024
	lru := NewCache(size)

	testType := func(c *Cache, t *testing.T) {
		t.Log("NewCache creates the correct type")
	}
	testType(lru, t)
}

func TestPush(t *testing.T) {
	lru := NewCache(1024)
	lru.Push("file.jpeg", []byte{1, 2, 3, 4, 5, 6, 7, 8})
	if len(*lru.heap) != 1 {
		t.Error("lru heap be 1 but instead it was", lru.heap)
	}
}

func TestGet(t *testing.T) {
	lru := NewCache(1024)

	inputBytes := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	lru.Push("file.jpeg", inputBytes)

	bytes, ok := lru.Get("file.jpeg")
	if !ok {
		t.Error("should have been able to get the file succesfully")
	}

	if !compare(bytes, inputBytes) {
		t.Error("should return the correct bytes")
	}

}
func TestGetMiss(t *testing.T) {
	lru := NewCache(1024)
	lru.Push("file.jpeg", []byte{1, 2, 3, 4, 5, 6, 7, 8})

	val, ok := lru.Get("Idon'tExist.jpeg.movie")
	if ok {
		t.Error("it should not be ok because the gotten file doesn't exist")
	}

	if val != nil {
		t.Error("it should be returning a nil value from lru.Get if it is a cache miss")
	}
}

func TestPushOrdersHeapCorrectly(t *testing.T) {
	lru := NewCache(1024)
	lru.Push("file.jpeg", []byte{1, 2, 3, 4, 5, 6, 7, 8})
	lru.Push("otherfile.jpeg", []byte{1, 2, 3, 4, 5, 6, 7, 8})

	heapVal := *lru.heap
	if heapVal[0].accessed.After(heapVal[1].accessed) {
		t.Error("the first element in the heap should be older than the second element")
	}
}

func TestPopsCorrectly(t *testing.T) {
	lru := NewCache(10)

	nineByteSlice := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
	oneByteSlice := []byte{0}

	lru.Push("niner", nineByteSlice)

	lru.Push("oner", oneByteSlice)

	if lru.currentSize != 10 {
		t.Error("lru should be counting the size correctly")
	}

	lru.Push("oner_2", oneByteSlice)

	if lru.currentSize != 2 {
		t.Error("should have pushed out the nine size element")
	}
}

func TestPushDuplicateFails(t *testing.T) {
	lru := NewCache(10)
	lru.Push("oner", []byte{1})
	lru.Push("oner", []byte{1})
	if lru.currentSize != 1 {
		t.Error("the currentSize state should be one")
	}

	if len(*lru.heap) != 1 {
		t.Error("the number of heap elements should be 1")
	}
}

func TestGetReordersHeap(t *testing.T) {
	lru := NewCache(10)

	lru.Push("oner", []byte{1})
	lru.Push("twoer", []byte{1})

	_, _ = lru.Get("oner")
	heapVal := *lru.heap
	if heapVal[0].name != "twoer" {
		t.Error("the first element (oldest) in the heap should be the last element accesed")
	}
}
