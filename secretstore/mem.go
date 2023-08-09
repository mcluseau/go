package secretstore

func memzero(ba []byte) {
	for i := range ba {
		ba[i] = 0
	}
}
