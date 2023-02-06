package cow

func Map[K comparable, T any](m *map[K]T) {
	dst := make(map[K]T, len(*m))
	for k, v := range *m {
		dst[k] = v
	}
	m = &dst
}

func MapSet[K comparable, T any](m *map[K]T, k K, v T) {
	Map(m)
	(*m)[k] = v
	return
}
