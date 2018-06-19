package byte_stream

type ByteStream struct {
	base         []byte
	dataOffset   int
	bufferOffset int
}

func (self *ByteStream) Read(buffer []byte) int {
	copy(buffer, self.GetData())
	return self.DiscardData(len(buffer))
}

func (self *ByteStream) DiscardData(dataSize int) int {
	if maxDataSize := self.GetDataSize(); dataSize > maxDataSize {
		dataSize = maxDataSize
	}

	if dataSize < 1 {
		return 0
	}

	self.dataOffset += dataSize

	if 2*self.dataOffset > self.bufferOffset {
		copy(self.base, self.GetData())
		self.bufferOffset -= self.dataOffset
		self.dataOffset = 0
	}

	return dataSize
}

func (self *ByteStream) Write(data []byte) {
	self.ReserveBuffer(len(data))
	copy(self.GetBuffer(), data)
	self.bufferOffset += len(data)
}

func (self *ByteStream) ReserveBuffer(bufferSize int) {
	if bufferSize < self.GetBufferSize() {
		return
	}

	if bufferSize < 1 {
		return
	}

	data := self.GetData()

	if bufferSize > len(self.base)-len(data) || 2*len(data) > len(self.base) {
		newBaseSize := int(nextPowerOfTwo(int64(self.bufferOffset + bufferSize)))
		self.base = make([]byte, newBaseSize)
	}

	copy(self.base, data)
	self.dataOffset = 0
	self.bufferOffset = len(data)
}

func (self *ByteStream) CommitBuffer(bufferSize int) int {
	if maxBufferSize := self.GetBufferSize(); bufferSize > maxBufferSize {
		bufferSize = maxBufferSize
	}

	if bufferSize < 1 {
		return 0
	}

	self.bufferOffset += bufferSize
	return bufferSize
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

func nextPowerOfTwo(number int64) int64 {
	number--
	number |= number >> 1
	number |= number >> 2
	number |= number >> 4
	number |= number >> 8
	number |= number >> 16
	number |= number >> 32
	number++
	return number
}
