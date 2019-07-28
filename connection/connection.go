package connection

import (
	"context"
	"net"
	"time"

	"github.com/let-z-go/toolkit/utils"
)

type Connection struct {
	raw     net.Conn
	preRWCs chan *connectionPreRWC
}

func (self *Connection) Init(raw net.Conn) *Connection {
	utils.Assert(raw != nil, func() string {
		return "toolkit/connection: invalid argument: raw is nil"
	})

	self.raw = raw
	self.preRWCs = make(chan *connectionPreRWC, 2)

	go func() {
		readCancellation := noCancellation
		readDeadline := time.Time{}
		writeCancellation := noCancellation
		writeDeadline := time.Time{}

		for {
			select {
			case preRWC := <-self.preRWCs:
				switch preRWC.type_ {
				case 'R':
					if !preRWC.deadline.Equal(readDeadline) {
						readDeadline = preRWC.deadline
						self.raw.SetReadDeadline(readDeadline)
					}

					readCancellation = preRWC.cancellation
					close(preRWC.completion)
				case 'W':
					if !preRWC.deadline.Equal(writeDeadline) {
						writeDeadline = preRWC.deadline
						self.raw.SetWriteDeadline(writeDeadline)
					}

					writeCancellation = preRWC.cancellation
					close(preRWC.completion)
				default: // case 'C':
					close(self.preRWCs)
					close(preRWC.completion)
					return
				}
			case <-readCancellation:
				readDeadline = time.Now()
				self.raw.SetReadDeadline(readDeadline)
				readCancellation = noCancellation
			case <-writeCancellation:
				writeDeadline = time.Now()
				self.raw.SetWriteDeadline(writeDeadline)
				writeCancellation = noCancellation
			}
		}
	}()

	return self
}

func (self *Connection) Close() error {
	completion := make(chan struct{})
	self.preRWCs <- &connectionPreRWC{'C', nil, time.Time{}, completion}
	<-completion
	raw := self.raw
	self.raw = nil
	return raw.Close()
}

func (self *Connection) Read(ctx context.Context, deadline time.Time, buffer []byte) (int, error) {
	self.PreRead(ctx, deadline)
	return self.DoRead(ctx, buffer)
}

func (self *Connection) PreRead(ctx context.Context, deadline time.Time) {
	completion := make(chan struct{})
	self.preRWCs <- &connectionPreRWC{'R', ctx.Done(), deadline, completion}
	<-completion
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
	completion := make(chan struct{})
	self.preRWCs <- &connectionPreRWC{'W', ctx.Done(), deadline, completion}
	<-completion
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

type connectionPreRWC struct {
	type_        byte
	cancellation <-chan struct{}
	deadline     time.Time
	completion   chan<- struct{}
}

var noCancellation = make(<-chan struct{})
