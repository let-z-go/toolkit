package utils

func MaxOfZero(x int64) int64 {
	return x &^ (x >> 15)
}
