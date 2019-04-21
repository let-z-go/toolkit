package utils

func Abs(x int64) int64 {
	y := x >> 63
	return (x ^ y) - y
}
