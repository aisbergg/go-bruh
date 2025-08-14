// Package util provides utility functions for various operations.
package util //nolint:revive

// DigitsInNumber returns the number of digits in a given integer.
func DigitsInNumber(n int) int {
	if n < 0 {
		n = -n
	}
	switch {
	case n < 10:
		return 1
	case n < 100:
		return 2
	case n < 1000:
		return 3
	case n < 10000:
		return 4
	case n < 100000:
		return 5
	case n < 1000000:
		return 6
	case n < 10000000:
		return 7
	case n < 100000000:
		return 8
	case n < 1000000000:
		return 9
	default:
		return 10
	}
}
