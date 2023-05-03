package memlog

import (
	"fmt"
	"sync"
	"testing"
)

func ExampleMemlog() {
	log := New[int]()

	stop := make(chan struct{})

	out := log.Subscribe(stop)

	const N = 3

	for n := 1; n <= N; n++ {
		log.Append(n)
	}

	for n := range out {
		fmt.Println(n)

		if n == N {
			// we've got our stop condition, get out and stop the subscription
			break
		}
	}

	close(stop)

	// Output:
	// 1
	// 2
	// 3
}

func ExampleNewWithBlockSize() {
	const blockSize = 10
	log := NewWithBlockSize[int](blockSize)

	stop := make(chan struct{})

	out := log.Subscribe(stop)

	const N = 3 * blockSize

	for n := 0; n < N; n++ {
		log.Append(n)
	}

	for n := range out {
		fmt.Printf("%02d", n)

		if n%blockSize == blockSize-1 {
			fmt.Println()
		} else {
			fmt.Print(" ")
		}

		if n == N-1 {
			// we've got our stop condition, get out and stop the subscription
			break
		}
	}
	fmt.Println()

	close(stop)

	// Output:
	// 00 01 02 03 04 05 06 07 08 09
	// 10 11 12 13 14 15 16 17 18 19
	// 20 21 22 23 24 25 26 27 28 29
}

func BenchmarkMemlogNoSub(b *testing.B) {
	benchNSub(b, 0)
}

func BenchmarkMemlog1Sub(b *testing.B) {
	benchNSub(b, 1)
}

func BenchmarkMemlog10Sub(b *testing.B) {
	benchNSub(b, 10)
}

func BenchmarkMemlog100Sub(b *testing.B) {
	benchNSub(b, 100)
}

func BenchmarkMemlog1000Sub(b *testing.B) {
	benchNSub(b, 1000)
}

func benchNSub(b *testing.B, nSub int) {
	log := New[int]()

	wg := sync.WaitGroup{}
	wg.Add(nSub)

	lastN := b.N - 1

	for subId := 0; subId < nSub; subId++ {
		// subId := subId

		stop := make(chan struct{})
		out := log.Subscribe(stop)

		go func() {
			for n := range out {
				// fmt.Println("sub=" + strconv.Itoa(subId) + " n=" + strconv.Itoa(n))
				if n == lastN {
					break
				}
			}

			close(stop)
			wg.Done()
		}()
	}

	for n := 0; n < b.N; n++ {
		log.Append(n)
	}

	wg.Wait()
}
