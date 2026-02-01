package a

import "fmt"

func badPanic() {
	panic("something went wrong") // want "panic should not be used in production code"
}

func goodPanic(x int) {
	if x < 0 {
		panic("x must be non-negative") // want "panic should not be used in production code"
	}
	fmt.Println(x)
}
