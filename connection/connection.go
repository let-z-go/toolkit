package connection

import (
	"context"
	"net"
	"time"
)

type Connection struct {
	raw     net.Conn
	preRWCs chan *connectionPreRWC
}

func (self *Connection) Adopt(raw net.Conn) {
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
					if preRWC.deadline != readDeadline {
						readDeadline = preRWC.deadline
						self.raw.SetReadDeadline(readDeadline)
					}

					readCancellation = preRWC.cancellation
					close(preRWC.completion)
				case 'W':
					if preRWC.deadline != writeDeadline {
						writeDeadline = preRWC.deadline
						self.raw.SetWriteDeadline(writeDeadline)
					}

					writeCancellation = preRWC.cancellation
					close(preRWC.completion)
				default: // case 'C':
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
}

func (self *Connection) Read(context_ context.Context, deadline time.Time, buffer []byte) (int, error) {
	self.PreRead(context_, deadline)
	return self.DoRead(context_, buffer)
}

func (self *Connection) PreRead(context_ context.Context, deadline time.Time) {
	completion := make(chan struct{})
	self.preRWCs <- &connectionPreRWC{'R', context_.Done(), deadline, completion}
	<-completion
}

func (self *Connection) DoRead(context_ context.Context, buffer []byte) (int, error) {
	n, e := self.raw.Read(buffer)

	if e != nil {
		if e2 := context_.Err(); e2 != nil {
			e = e2
		}
	}

	return n, e
}

func (self *Connection) Write(context_ context.Context, deadline time.Time, data []byte) (int, error) {
	self.PreWrite(context_, deadline)
	return self.DoWrite(context_, data)
}

func (self *Connection) PreWrite(context_ context.Context, deadline time.Time) {
	completion := make(chan struct{})
	self.preRWCs <- &connectionPreRWC{'W', context_.Done(), deadline, completion}
	<-completion
}

func (self *Connection) DoWrite(context_ context.Context, data []byte) (int, error) {
	n, e := self.raw.Write(data)

	if e != nil {
		if e2 := context_.Err(); e2 != nil {
			e = e2
		}
	}

	return n, e
}

func (self *Connection) Close() error {
	completion := make(chan struct{})
	self.preRWCs <- &connectionPreRWC{'C', nil, time.Time{}, completion}
	<-completion
	e := self.raw.Close()
	self.raw = nil
	self.preRWCs = nil
	return e
}

type connectionPreRWC struct {
	type_        byte
	cancellation <-chan struct{}
	deadline     time.Time
	completion   chan<- struct{}
}

var noCancellation = make(<-chan struct{})
