package watchable

import (
	"context"
	"time"
)

type Watch[T any] struct {
	w      *Watchable[T]
	rev    uint64
	stopCh chan struct{}
}

func (w *Watchable[T]) NewWatch() *Watch[T] {
	return &Watch[T]{w: w, rev: 0, stopCh: make(chan struct{}, 1)}
}
func (w *Watchable[T]) NewWatchCh() (ch <-chan T, stop func()) {
	watch := w.NewWatch()
	return watch.Chan(), watch.Stop
}

func (w *Watchable[T]) NewWatchWithContext(ctx context.Context) (watch *Watch[T]) {
	watch = &Watch[T]{w: w, rev: 0, stopCh: make(chan struct{}, 1)}

	go func() {
		select {
		case <-ctx.Done():
			watch.Stop()
		case <-watch.stopCh:
			// skip
		}
	}()

	return
}

func (w *Watch[T]) Stop() {
	ch := w.stopCh
	if ch != nil {
		close(ch)
		w.stopCh = nil
	}
}

func (w *Watch[T]) NextWithTimeout(timeout time.Duration) (next T, ok, timedOut bool) {
	stopCh := w.stopCh

	if timeout != 0 {
		stopCh = make(chan struct{}, 1)

		go func() {
			select {
			case <-w.stopCh:
				// noop
			case <-time.After(timeout):
				// noop
			}
			close(stopCh)
		}()
	}

	next, rev, timedOut := w.w.NextWithTimeout(w.rev, stopCh)

	if timedOut {
		return
	}

	if rev == 0 {
		// closed
		ok = false
		return
	}

	w.rev = rev
	ok = true
	return
}

func (w *Watch[T]) Next() (next T, ok bool) {
	next, ok, _ = w.NextWithTimeout(0)
	return
}

func (w *Watch[T]) NextNoValue() (ok bool) {
	_, ok = w.Next()
	return
}

func (w *Watch[T]) Chan() <-chan T {
	ch := make(chan T, 1)
	go func() {
		for {
			v, ok := w.Next()
			if !ok {
				close(ch)
				return
			}
			ch <- v
		}
	}()
	return ch
}
