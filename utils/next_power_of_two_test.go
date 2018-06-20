package utils

import (
	"testing"
)

func TestNextPowerOfTwo(t *testing.T) {
	if x := NextPowerOfTwo(0); x != 0 {
		t.Errorf("%#v", x)
	}

	if x := NextPowerOfTwo(1); x != 1 {
		t.Errorf("%#v", x)
	}

	if x := NextPowerOfTwo(2); x != 2 {
		t.Errorf("%#v", x)
	}

	if x := NextPowerOfTwo(3); x != 4 {
		t.Errorf("%#v", x)
	}
}
