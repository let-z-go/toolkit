package utils

import (
	"math"
	"testing"
)

func TestMaxOfZero(t *testing.T) {
	if x := MaxOfZero(-199); x != 0 {
		t.Errorf("%#v", x)
	}

	if x := MaxOfZero(-1); x != 0 {
		t.Errorf("%#v", x)
	}

	if x := MaxOfZero(0); x != 0 {
		t.Errorf("%#v", x)
	}

	if x := MaxOfZero(100); x != 100 {
		t.Errorf("%#v", x)
	}

	if x := MaxOfZero(math.MaxInt64); x != math.MaxInt64 {
		t.Errorf("%#v", x)
	}
}
