package connection

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/let-z-go/toolkit/utils"
)

type Connection struct {
	raw                     net.Conn
	preReading              chan struct{}
	preWriting              chan struct{}
	lockOfPreReading        sync.Mutex
	lockOfPreWriting        sync.Mutex
	latestReadCancellation  <-chan struct{}
	latestWriteCancellation <-chan struct{}
}

func (self *Connection) Init(raw net.Conn) *Connection {
	utils.Assert(raw != nil, func() string {
		return "toolkit/connection: invalid argument: raw is nil"
	})

	self.raw = raw
	self.preReading = make(chan struct{}, 1)
	self.preWriting = make(chan struct{}, 1)

	go func() {
		readCancellation := noCancellation
		writeCancellation := noCancellation

		for {
			select {
			case _, ok := <-self.preReading:
				if !ok {
					return
				}

				self.lockOfPreReading.Lock()
				readCancellation = self.latestReadCancellation
				self.lockOfPreReading.Unlock()
			case _, ok := <-self.preWriting:
				if !ok {
					return
				}

				self.lockOfPreWriting.Lock()
				writeCancellation = self.latestWriteCancellation
				self.lockOfPreWriting.Unlock()
			case <-readCancellation:
				self.lockOfPreReading.Lock()

				if readCancellation == self.latestReadCancellation {
					raw.SetReadDeadline(time.Now())
				}

				self.lockOfPreReading.Unlock()
				readCancellation = noCancellation
			case <-writeCancellation:
				self.lockOfPreWriting.Lock()

				if writeCancellation == self.latestWriteCancellation {
					raw.SetWriteDeadline(time.Now())
				}

				self.lockOfPreWriting.Unlock()
				writeCancellation = noCancellation
			}
		}
	}()

	return self
}

func (self *Connection) Close() error {
	close(self.preReading)
	close(self.preWriting)
	raw := self.raw
	self.raw = nil
	return raw.Close()
}

func (self *Connection) Read(ctx context.Context, deadline time.Time, buffer []byte) (int, error) {
	self.PreRead(ctx, deadline)
	return self.DoRead(ctx, buffer)
}

func (self *Connection) PreRead(ctx context.Context, deadline time.Time) {
	self.lockOfPreReading.Lock()
	self.latestReadCancellation = ctx.Done()
	self.raw.SetReadDeadline(deadline)
	self.lockOfPreReading.Unlock()

	select {
	case self.preReading <- struct{}{}:
	default:
	}
}

func (self *Connection) DoRead(ctx context.Context, buffer []byte) (int, error) {
	n, err := self.raw.Read(buffer)

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
	self.latestWriteCancellation = ctx.Done()
	self.raw.SetWriteDeadline(deadline)
	self.lockOfPreWriting.Unlock()

	select {
	case self.preWriting <- struct{}{}:
	default:
	}
}

func (self *Connection) DoWrite(ctx context.Context, data []byte) (int, error) {
	n, err := self.raw.Write(data)

	if err != nil {
		if err2 := ctx.Err(); err2 != nil {
			err = err2
		}
	}

	return n, err
}

func (self *Connection) IsClosed() bool {
	return self.raw == nil
}

var noCancellation = make(<-chan struct{})
