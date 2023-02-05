package watchable

import (
	"sync"
)

type Watchable[T any] struct {
	OnChange func(*T)

	clone func(T) T

	v      T
	l      *sync.RWMutex
	c      *sync.Cond
	rev    uint64
	closed bool
}

func New[T any]() *Watchable[T] {
	return NewWithClone[T](nil)
}

func NewClonable[T interface{ Clone() T }]() *Watchable[T] {
	return NewWithClone(func(v T) T { return v.Clone() })
}

func NewWithClone[T any](clone func(T) T) *Watchable[T] {
	l := new(sync.RWMutex)
	return &Watchable[T]{
		l:     l,
		c:     sync.NewCond(l),
		clone: clone,
	}
}

func (w *Watchable[T]) NextWithTimeout(rev uint64, stopCh <-chan struct{}) (v T, nextRev uint64, timedOut bool) {
	runCh := make(chan struct{}, 1)
	defer close(runCh)

	if stopCh != nil {
		go func() {
			select {
			case <-stopCh:
				timedOut = true
				w.c.Broadcast()

			case <-runCh:
				// noop
			}
		}()
	}

	w.c.L.Lock()
	defer w.c.L.Unlock()

	for {
		if w.rev > rev {
			break
		}
		if w.closed {
			return
		}
		if timedOut {
			v = w.v // return the current value anyway
			return
		}
		w.c.Wait()
	}

	v = w.v
	nextRev = w.rev

	return
}

func (w *Watchable[T]) Next(rev uint64) (v T, nextRev uint64) {
	v, nextRev, _ = w.NextWithTimeout(rev, nil)
	return
}

func (w *Watchable[T]) View(view func(v T)) (rev uint64) {
	w.l.RLock()
	defer w.l.RUnlock()

	view(w.v)

	return w.rev
}

func (w *Watchable[T]) Get() (v T) {
	v, _ = w.GetWithRev()
	return
}

func (w *Watchable[T]) GetWithRev() (v T, rev uint64) {
	w.c.L.Lock()

	v = w.v
	rev = w.rev

	w.c.L.Unlock()
	return
}

func (w *Watchable[T]) Set(v T) {
	w.c.L.Lock()

	w.v = v
	w.rev++

	w.c.L.Unlock()
	w.c.Broadcast()
}

func (w *Watchable[T]) Close() {
	w.c.L.Lock()
	w.closed = true
	w.c.L.Unlock()
	w.c.Broadcast()
}

func (w *Watchable[T]) Change(change func(v *T)) {
	w.Update(func(v T) (newV T, changed bool) {
		if w.clone == nil {
			newV = v
		} else {
			newV = w.clone(v)
		}
		change(&newV)
		changed = true
		return
	})
}

func (w *Watchable[T]) Update(update func(v T) (newV T, changed bool)) {
	w.c.L.Lock()

	v, changed := update(w.v)
	if !changed {
		w.c.L.Unlock()
		return
	}

	if onChange := w.OnChange; onChange != nil {
		onChange(&v)
	}

	w.v = v
	w.rev++

	w.c.L.Unlock()
	w.c.Broadcast()
}
