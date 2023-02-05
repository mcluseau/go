package watchable

func IntIncrement(x int) (y int, changed bool) { return x+1, true }
func IntDecrement(x int) (y int, changed bool) { return x-1, true }

