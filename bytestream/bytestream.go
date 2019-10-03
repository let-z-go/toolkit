package bytestream

import (
	"github.com/let-z-go/toolkit/utils"
)

type ByteStream struct {
	base         []byte
	dataOffset   int
	bufferOffset int
}

func (self *ByteStream) Read(buffer []byte) int {
	dataSize := copy(buffer, self.GetData())
	self.doSkip(dataSize)
	return dataSize
}

func (self *ByteStream) Skip(dataSize int) int {
	dataSize = int(utils.MaxOfZero(int64(dataSize)))

	if maxDataSize := self.GetDataSize(); dataSize > maxDataSize {
		dataSize = maxDataSize
	}

	self.doSkip(dataSize)
	return dataSize
}

func (self *ByteStream) Write(data []byte) {
	self.ReserveBuffer(len(data))
	copy(self.GetBuffer(), data)
	self.doCommitBuffer(len(data))
}

func (self *ByteStream) WriteDirectly(bufferSize int, callback func([]byte) error) error {
	bufferSize = int(utils.MaxOfZero(int64(bufferSize)))
	self.ReserveBuffer(bufferSize)

	if err := callback(self.GetBuffer()); err != nil {
		return err
	}

	self.doCommitBuffer(bufferSize)
	return nil
}

func (self *ByteStream) Unwrite(dataSize int) int {
	dataSize = int(utils.MaxOfZero(int64(dataSize)))

	if maxDataSize := self.GetDataSize(); dataSize > maxDataSize {
		dataSize = maxDataSize
	}

	self.bufferOffset -= dataSize

	if self.dataOffset >= self.bufferOffset/2 {
		self.setData(self.GetData())
	}

	return dataSize
}

func (self *ByteStream) ReserveBuffer(bufferSize int) {
	bufferSize = int(utils.MaxOfZero(int64(bufferSize)))

	if self.GetBufferSize() >= bufferSize {
		return
	}

	data := self.GetData()

	if len(self.base)-len(data) < bufferSize {
		self.base = make([]byte, int(utils.NextPowerOfTwo(int64(len(data)+bufferSize))))
	} else {
		if len(data) > len(self.base)/2 {
			self.base = make([]byte, 2*len(self.base))
		}
	}

	self.setData(data)
}

func (self *ByteStream) CommitBuffer(bufferSize int) int {
	bufferSize = int(utils.MaxOfZero(int64(bufferSize)))

	if maxBufferSize := self.GetBufferSize(); bufferSize > maxBufferSize {
		bufferSize = maxBufferSize
	}

	self.doCommitBuffer(bufferSize)
	return bufferSize
}

func (self *ByteStream) Expand() {
	data := self.GetData()
	newSize := len(self.base) * 2
	self.base = make([]byte, newSize)
	self.setData(data)
}

func (self *ByteStream) Shrink(minSize int) {
	data := self.GetData()

	if minSize < len(data) {
		minSize = len(data)
	}

	minSize = int(utils.NextPowerOfTwo(int64(minSize)))

	if len(self.base) == minSize {
		return
	}

	self.base = make([]byte, minSize)
	self.setData(data)
}

func (self *ByteStream) GetData() []byte {
	return self.base[self.dataOffset:self.bufferOffset]
}

func (self *ByteStream) GetDataSize() int {
	return self.bufferOffset - self.dataOffset
}

func (self *ByteStream) GetBuffer() []byte {
	return self.base[self.bufferOffset:]
}

func (self *ByteStream) GetBufferSize() int {
	return len(self.base) - self.bufferOffset
}

func (self *ByteStream) Size() int {
	return len(self.base)
}

func (self *ByteStream) doSkip(dataSize int) {
	self.dataOffset += dataSize

	if self.dataOffset >= self.bufferOffset/2 {
		self.setData(self.GetData())
	}
}

func (self *ByteStream) setData(data []byte) {
	copy(self.base, data)
	self.dataOffset = 0
	self.bufferOffset = len(data)
}

func (self *ByteStream) doCommitBuffer(bufferSize int) {
	self.bufferOffset += bufferSize
}
