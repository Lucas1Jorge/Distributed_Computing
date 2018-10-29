package main

import "fmt"

// fibonacci is a function that returns
// a function that returns an int.
func fibonacci() func() int {
	fib_1, fib_2 := 1, 0
	return func() int {
		aux := fib_2
		fib_2 = fib_1
		fib_1 = aux + fib_1
		
		return fib_2
	}
}

func main() {
	f := fibonacci()
	for i := 0; i < 10; i++ {
		fmt.Println(f())
	}
}
