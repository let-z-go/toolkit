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

func (self *Connection) Init(underlying net.Conn) *Connection {
	utils.Assert(underlying != nil, func() string {
		return "toolkit/connection: invalid argument: underlying is nil"
	})

	self.underlying = underlying
	self.preReading = make(chan struct{}, 1)
	self.preWriting = make(chan struct{}, 1)

	go func() {
		readCtx := context.Background()
		writeCtx := context.Background()

		for {
			select {
			case _, ok := <-self.preReading:
				if !ok {
					return
				}

				self.lockOfPreReading.Lock()
				readCtx = self.latestReadCtx
				self.lockOfPreReading.Unlock()
			case _, ok := <-self.preWriting:
				if !ok {
					return
				}

				self.lockOfPreWriting.Lock()
				writeCtx = self.latestWriteCtx
				self.lockOfPreWriting.Unlock()
			case <-readCtx.Done():
				self.lockOfPreReading.Lock()

				if readCtx == self.latestReadCtx {
					underlying.SetReadDeadline(time.Now())
				}

				self.lockOfPreReading.Unlock()
				readCtx = context.Background()
			case <-writeCtx.Done():
				self.lockOfPreWriting.Lock()

				if writeCtx == self.latestWriteCtx {
					underlying.SetWriteDeadline(time.Now())
				}

				self.lockOfPreWriting.Unlock()
				writeCtx = context.Background()
			}
		}
	}()

	return self
}

func (self *Connection) Close() error {
	close(self.preReading)
	close(self.preWriting)
	underlying := self.underlying
	self.underlying = nil
	return underlying.Close()
}

func (self *Connection) Read(ctx context.Context, deadline time.Time, buffer []byte) (int, error) {
	self.PreRead(ctx, deadline)
	return self.DoRead(ctx, buffer)
}

func (self *Connection) PreRead(ctx context.Context, deadline time.Time) {
	self.lockOfPreReading.Lock()
	self.latestReadCtx = ctx
	self.underlying.SetReadDeadline(deadline)
	self.lockOfPreReading.Unlock()

	select {
	case self.preReading <- struct{}{}:
	default:
	}
}

func (self *Connection) DoRead(ctx context.Context, buffer []byte) (int, error) {
	n, err := self.underlying.Read(buffer)

	if err != nil {
		if err2 := ctx.Err(); err2 != nil {
			err = err2
		}
	}

	return n, err
}

func (self *Connection) Write(ctx context.Context, deadline time.Time, data []byte) (int, error) {
	self.PreWrite(ctx, deadline)
	return self.DoWrite(ctx, data)
}

func (self *Connection) PreWrite(ctx context.Context, deadline time.Time) {
	self.lockOfPreWriting.Lock()
	self.latestWriteCtx = ctx
	self.underlying.SetWriteDeadline(deadline)
	self.lockOfPreWriting.Unlock()

	select {
	case self.preWriting <- struct{}{}:
	default:
	}
}

func (self *Connection) DoWrite(ctx context.Context, data []byte) (int, error) {
	n, err := self.underlying.Write(data)

	if err != nil {
		if err2 := ctx.Err(); err2 != nil {
			err = err2
		}
	}

	return n, err
}

func (self *Connection) IsClosed() bool {
	return self.underlying == nil
}
