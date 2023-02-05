package watchable

func UpdateIfChanged[T comparable](w *Watchable[T], newValue T) {
	w.Update(func(prevV T) (newV T, changed bool) {
		if prevV == newValue {
			return
		}
		return newValue, true
	})
}
