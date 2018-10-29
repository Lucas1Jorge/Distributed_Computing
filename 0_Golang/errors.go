package main

import (
	"fmt"
	// "strconv"
)

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

type ErrNegativeSqrt struct {
	Val float64
}

func (e ErrNegativeSqrt) Error() string {
	return fmt.Sprintf("cannot Sqrt negative number: %v", e.Val)
}

func Sqrt(x float64) (float64, error) {
	if x < 0 {
		var e ErrNegativeSqrt
		e.Val = x

		return 0, e
	}
	
	z := 1.0

	for abs(z*z - x) > 1e-5 {
		z -= (z*z - x) / (2 * z)
		// fmt.Println(z)
	}

	return z, nil
}

func main() {
	fmt.Println(Sqrt(2))
	fmt.Println(Sqrt(-2))
}
