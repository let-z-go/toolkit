package utils

import (
	"math"
	"testing"
)

func TestAbs(t *testing.T) {
	if n := Abs(0); n != 0 {
		t.Errorf("%#v", n)
	}
	if n := Abs(100); n != 100 {
		t.Errorf("%#v", n)
	}
	if n := Abs(-100); n != 100 {
		t.Errorf("%#v", n)
	}
	if n := Abs(math.MaxInt64); n != math.MaxInt64 {
		t.Errorf("%#v", n)
	}
	if n := Abs(-math.MaxInt64); n != math.MaxInt64 {
		t.Errorf("%#v", n)
	}
	if n := Abs(math.MinInt64); n != math.MinInt64 {
		t.Errorf("%#v", n)
	}
}
