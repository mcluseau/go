package cow

func Slice[T any](a *[]T) {
	dst := make([]T, len(*a))
	copy(dst, *a)
	a = &dst
}

func SliceSet[T any](a *[]T, idx int, v T) {
	Slice(a)
	(*a)[idx] = v
	return
}
