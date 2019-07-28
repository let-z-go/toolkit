package utils

import (
	"errors"
)

func Assert(ok bool, messageFactory func() string) {
	if !ok {
		panic(errors.New(messageFactory()))
	}
}
