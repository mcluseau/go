package watchable

func Propagate[S any, T any](stop <-chan struct{}, wSrc *Watchable[S], wDst *Watchable[T], propagate func(S, *T)) {
	srcCh, srcStop := wSrc.NewWatchCh()
	defer srcStop()

	for src := range srcCh {
		wDst.Change(func(dst *T) { propagate(src, dst) })
	}
}
