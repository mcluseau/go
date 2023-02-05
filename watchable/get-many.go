package watchable

// GetMany gets a consistent snapshot of many watchables at one
func GetMany[T any](watchables []*Watchable[T]) (ret []T) {
	ret = make([]T, len(watchables))

	for _, w := range watchables {
		w.l.RLock()
	}

	defer func() {
		for _, w := range watchables {
			w.l.RUnlock()
		}
	}()

	for i, w := range watchables {
		ret[i] = w.v
	}

	return
}
