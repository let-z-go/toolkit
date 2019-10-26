package bytestream

import (
	"github.com/let-z-go/toolkit/utils"
)

type ByteStream struct {
	base         []byte
	dataOffset   int
	bufferOffset int
}

func (bs *ByteStream) Read(buffer []byte) int {
	dataSize := copy(buffer, bs.GetData())
	bs.doSkip(dataSize)
	return dataSize
}

func (bs *ByteStream) Skip(dataSize int) int {
	dataSize = int(utils.MaxOfZero(int64(dataSize)))

	if maxDataSize := bs.GetDataSize(); dataSize > maxDataSize {
		dataSize = maxDataSize
	}

	bs.doSkip(dataSize)
	return dataSize
}

func (bs *ByteStream) Write(data []byte) {
	bs.ReserveBuffer(len(data))
	copy(bs.GetBuffer(), data)
	bs.doCommitBuffer(len(data))
}

func (bs *ByteStream) WriteDirectly(bufferSize int, callback func([]byte) error) error {
	bufferSize = int(utils.MaxOfZero(int64(bufferSize)))
	bs.ReserveBuffer(bufferSize)

	if err := callback(bs.GetBuffer()); err != nil {
		return err
	}

	bs.doCommitBuffer(bufferSize)
	return nil
}

func (bs *ByteStream) Unwrite(dataSize int) int {
	dataSize = int(utils.MaxOfZero(int64(dataSize)))

	if maxDataSize := bs.GetDataSize(); dataSize > maxDataSize {
		dataSize = maxDataSize
	}

	bs.bufferOffset -= dataSize

	if bs.dataOffset*2 >= bs.bufferOffset {
		bs.setData(bs.GetData())
	}

	return dataSize
}

func (bs *ByteStream) ReserveBuffer(bufferSize int) {
	bufferSize = int(utils.MaxOfZero(int64(bufferSize)))

	if bs.GetBufferSize() >= bufferSize {
		return
	}

	data := bs.GetData()

	if len(bs.base)-len(data) < bufferSize {
		bs.base = make([]byte, int(utils.NextPowerOfTwo(int64(len(data)+bufferSize))))
	} else {
		if len(data)*2 > len(bs.base) {
			bs.base = make([]byte, 2*len(bs.base))
		}
	}

	bs.setData(data)
}

func (bs *ByteStream) CommitBuffer(bufferSize int) int {
	bufferSize = int(utils.MaxOfZero(int64(bufferSize)))

	if maxBufferSize := bs.GetBufferSize(); bufferSize > maxBufferSize {
		bufferSize = maxBufferSize
	}

	bs.doCommitBuffer(bufferSize)
	return bufferSize
}

func (bs *ByteStream) Expand() {
	data := bs.GetData()
	newSize := len(bs.base) * 2
	bs.base = make([]byte, newSize)
	bs.setData(data)
}

func (bs *ByteStream) Shrink(minSize int) {
	data := bs.GetData()

	if minSize < len(data) {
		minSize = len(data)
	}

	minSize = int(utils.NextPowerOfTwo(int64(minSize)))

	if len(bs.base) == minSize {
		return
	}

	bs.base = make([]byte, minSize)
	bs.setData(data)
}

func (bs *ByteStream) GetData() []byte {
	return bs.base[bs.dataOffset:bs.bufferOffset]
}

func (bs *ByteStream) GetDataSize() int {
	return bs.bufferOffset - bs.dataOffset
}

func (bs *ByteStream) GetBuffer() []byte {
	return bs.base[bs.bufferOffset:]
}

func (bs *ByteStream) GetBufferSize() int {
	return len(bs.base) - bs.bufferOffset
}

func (bs *ByteStream) Size() int {
	return len(bs.base)
}

func (bs *ByteStream) doSkip(dataSize int) {
	bs.dataOffset += dataSize

	if bs.dataOffset*2 >= bs.bufferOffset {
		bs.setData(bs.GetData())
	}
}

func (bs *ByteStream) setData(data []byte) {
	copy(bs.base, data)
	bs.dataOffset = 0
	bs.bufferOffset = len(data)
}

func (bs *ByteStream) doCommitBuffer(bufferSize int) {
	bs.bufferOffset += bufferSize
}
