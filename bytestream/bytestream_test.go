package bytestream

import (
	"bytes"
	"testing"
)

func TestReadWriteByteStream1(t *testing.T) {
	var bs ByteStream
	bs.ReserveBuffer(10)

	if bufsz := bs.GetBufferSize(); bufsz != 16 {
		t.Errorf("%#v", bufsz)
	}

	bs.Write([]byte("12345"))
	var b1 [1]byte
	var b2 [2]byte
	var b3 [3]byte

	if n := bs.Read(b2[:]); n != len(b2) {
		t.Errorf("%#v", n)
	}

	if !bytes.Equal(b2[:], []byte("12")) {
		t.Errorf("%#v", b2)
	}

	if bufsz := bs.GetBufferSize(); bufsz != 11 {
		t.Errorf("%#v", bufsz)
	}

	if n := bs.Read(b1[:]); n != len(b1) {
		t.Errorf("%#v", n)
	}

	if !bytes.Equal(b1[:], []byte("3")) {
		t.Errorf("%#v", b2)
	}

	if bufsz := bs.GetBufferSize(); bufsz != 14 {
		t.Errorf("%#v", bufsz)
	}

	if n := bs.Read(b3[:]); n != 2 {
		t.Errorf("%#v", n)
	}

	if !bytes.Equal(b3[:], []byte("45\x00")) {
		t.Errorf("%#v", b2)
	}

	if bufsz := bs.GetBufferSize(); bufsz != 16 {
		t.Errorf("%#v", bufsz)
	}
}

func TestReadWriteByteStream2(t *testing.T) {
	var bs ByteStream
	bs.ReserveBuffer(10)
	bs.Write([]byte("0123456789"))
	bs.Skip(3)

	if bufsz := bs.GetBufferSize(); bufsz != 6 {
		t.Errorf("%#v", bufsz)
	}

	bs.ReserveBuffer(8)

	if bufsz := bs.GetBufferSize(); bufsz != 9 {
		t.Errorf("%#v", bufsz)
	}

	bs.Write([]byte("012"))

	if d := bs.GetData(); !bytes.Equal(d, []byte("3456789012")) {
		t.Errorf("%#v", d)
	}
}

func TestReadWriteByteStream3(t *testing.T) {
	var bs ByteStream
	bs.ReserveBuffer(10)
	bs.Write([]byte("0123456789"))
	bs.Skip(3)
	bs.ReserveBuffer(10)

	if bufsz := bs.GetBufferSize(); bufsz != 25 {
		t.Errorf("%#v", bufsz)
	}
}
