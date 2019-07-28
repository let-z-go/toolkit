package utils

import (
	"context"
	"testing"
	"time"
)

func TestDelayContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t0 := time.Now()
	ctx2 := DelayContext(ctx, time.Second)
	cancel()
	<-ctx2.Done()
	if d := time.Since(t0); d < time.Second {
		t.Fatalf("d=%#v<%#v", d, time.Second)
	}
}
