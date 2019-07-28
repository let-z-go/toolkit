package utils

import (
	"fmt"
	"testing"
)

func TestAssert(t *testing.T) {
	defer func() {
		if r := recover(); fmt.Sprintf("%s", r) != "2" {
			t.Fatalf("r=%#v", r)
		}
	}()
	Assert(1 == 1, func() string {
		return "1"
	})
	Assert(2 != 2, func() string {
		return "2"
	})
}
