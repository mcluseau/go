package watchable

import "fmt"

func ExampleUpdateIfChanged() {
	w := New[float64]()

	w.Set(1)
	fmt.Println("rev", w.rev)

	UpdateIfChanged(w, 1)
	fmt.Println("rev", w.rev)

	UpdateIfChanged(w, 2)
	fmt.Println("rev", w.rev)

	UpdateIfChanged(w, 2)
	fmt.Println("rev", w.rev)

	UpdateIfChanged(w, 3)
	fmt.Println("rev", w.rev)

	// Output:
	// rev 1
	// rev 1
	// rev 2
	// rev 2
	// rev 3
}
