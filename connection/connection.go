package connection

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/let-z-go/toolkit/utils"
)

type Connection struct {
	underlying       net.Conn
	preReading       chan struct{}
	preWriting       chan struct{}
	lockOfPreReading sync.Mutex
	lockOfPreWriting sync.Mutex
	latestReadCtx    context.Context
	latestWriteCtx   context.Context
}

func (c *Connection) Init(underlying net.Conn) *Connection {
	utils.Assert(underlying != nil, func() string {
		return "toolkit/connection: invalid argument: underlying is nil"
	})

	c.underlying = underlying
	c.preReading = make(chan struct{}, 1)
	c.preWriting = make(chan struct{}, 1)

	go func() {
		readCtx := context.Background()
		writeCtx := context.Background()

		for {
			select {
			case _, ok := <-c.preReading:
				if !ok {
					return
				}

				c.lockOfPreReading.Lock()
				readCtx = c.latestReadCtx
				c.lockOfPreReading.Unlock()
			case _, ok := <-c.preWriting:
				if !ok {
					return
				}

				c.lockOfPreWriting.Lock()
				writeCtx = c.latestWriteCtx
				c.lockOfPreWriting.Unlock()
			case <-readCtx.Done():
				c.lockOfPreReading.Lock()

				if readCtx == c.latestReadCtx {
					underlying.SetReadDeadline(time.Now())
				}

				c.lockOfPreReading.Unlock()
				readCtx = context.Background()
			case <-writeCtx.Done():
				c.lockOfPreWriting.Lock()

				if writeCtx == c.latestWriteCtx {
					underlying.SetWriteDeadline(time.Now())
				}

				c.lockOfPreWriting.Unlock()
				writeCtx = context.Background()
			}
		}
	}()

	return c
}

func (c *Connection) Close() error {
	close(c.preReading)
	close(c.preWriting)
	underlying := c.underlying
	c.underlying = nil
	return underlying.Close()
}

func (c *Connection) Read(ctx context.Context, deadline time.Time, buffer []byte) (int, error) {
	c.PreRead(ctx, deadline)
	return c.DoRead(ctx, buffer)
}

func (c *Connection) PreRead(ctx context.Context, deadline time.Time) {
	c.lockOfPreReading.Lock()
	c.latestReadCtx = ctx
	c.underlying.SetReadDeadline(deadline)
	c.lockOfPreReading.Unlock()

	select {
	case c.preReading <- struct{}{}:
	default:
	}
}

func (c *Connection) DoRead(ctx context.Context, buffer []byte) (int, error) {
	n, err := c.underlying.Read(buffer)

	if err != nil {
		if err2 := ctx.Err(); err2 != nil {
			err = err2
		}
	}

	return n, err
}

func (c *Connection) Write(ctx context.Context, deadline time.Time, data []byte) (int, error) {
	c.PreWrite(ctx, deadline)
	return c.DoWrite(ctx, data)
}

func (c *Connection) PreWrite(ctx context.Context, deadline time.Time) {
	c.lockOfPreWriting.Lock()
	c.latestWriteCtx = ctx
	c.underlying.SetWriteDeadline(deadline)
	c.lockOfPreWriting.Unlock()

	select {
	case c.preWriting <- struct{}{}:
	default:
	}
}

func (c *Connection) DoWrite(ctx context.Context, data []byte) (int, error) {
	n, err := c.underlying.Write(data)

	if err != nil {
		if err2 := ctx.Err(); err2 != nil {
			err = err2
		}
	}

	return n, err
}

func (c *Connection) IsClosed() bool {
	return c.underlying == nil
}
