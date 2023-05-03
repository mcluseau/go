package memlog

import (
	"reflect"
	"sync"
)

type Memlog[T any] struct {
	l           sync.Mutex
	tail        *entryBlock[T]
	tailNextPos int
}

type entryBlock[T any] struct {
	entries   []entry[T]
	nextSetCh chan struct{}
	next      *entryBlock[T]
}

type entry[T any] struct {
	value T
	setCh chan struct{}
}

func New[T any]() (log *Memlog[T]) {
	// guess a block size around 4kiB, or 16 entries minimum
	var entry entry[T]
	entrySize := reflect.TypeOf(entry).Size()
	blockSize := 4096 / int(entrySize)
	if blockSize < 16 {
		blockSize = 16
	}

	return NewWithBlockSize[T](blockSize)
}

func NewWithBlockSize[T any](blockSize int) (log *Memlog[T]) {
	log = &Memlog[T]{
		tail: newEntryBlock[T](blockSize),
	}
	return
}

func newEntryBlock[T any](size int) *entryBlock[T] {
	block := entryBlock[T]{
		entries:   make([]entry[T], size),
		nextSetCh: make(chan struct{}),
	}

	for i := range block.entries {
		block.entries[i].setCh = make(chan struct{})
	}

	return &block
}

func (log *Memlog[T]) Append(value T) {
	log.l.Lock()

	if log.tailNextPos == len(log.tail.entries) {
		newBlock := newEntryBlock[T](len(log.tail.entries))

		log.tail.next = newBlock
		close(log.tail.nextSetCh)

		log.tail = newBlock
		log.tailNextPos = 0
	}

	log.tail.entries[log.tailNextPos].value = value
	close(log.tail.entries[log.tailNextPos].setCh)

	log.tailNextPos++

	log.l.Unlock()
}

func (log *Memlog[T]) Subscribe(stop <-chan struct{}) <-chan T {
	log.l.Lock()
	tail := log.tail
	tailPos := log.tailNextPos
	log.l.Unlock()

	out := make(chan T) // no buffering needed

	go func() {
		defer close(out)

		for {
			if tailPos == len(tail.entries) {
				// if at the end of the current block, move to the next one
				select {
				case _, _ = <-tail.nextSetCh:
					tail = tail.next
					tailPos = 0

				case _, _ = <-stop:
					return
				}
			}

			// consume block
			select {
			case _, _ = <-tail.entries[tailPos].setCh:
				select {
				case out <- tail.entries[tailPos].value:
					tailPos++

				case _, _ = <-stop:
					return
				}

			case _, _ = <-stop:
				return
			}
		}
	}()

	return out
}
