package logger

import (
	"sync"
)

const bufferMaxEntries = 1000

var buffer = newCircularBuffer(bufferMaxEntries)

type circularBuffer struct {
	entries []entry
	index   int
	size    int
	isFull  bool
	mu      sync.RWMutex
}

type entry struct {
	Id        uint64 `json:"id"`
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}

func newCircularBuffer(size int) *circularBuffer {
	return &circularBuffer{
		entries: make([]entry, size),
		size:    size,
	}
}

func (b *circularBuffer) add(entry entry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries[b.index] = entry
	b.index = (b.index + 1) % b.size
	if b.index == 0 && !b.isFull {
		b.isFull = true
	}
}

func (b *circularBuffer) getAll() []entry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var entries []entry
	if b.isFull {
		entries = make([]entry, b.size)
		copy(entries, b.entries[b.index:])
		copy(entries[b.size-b.index:], b.entries[:b.index])
	} else {
		entries = make([]entry, b.index)
		copy(entries, b.entries[:b.index])
	}
	return entries
}
