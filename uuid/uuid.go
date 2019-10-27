package uuid

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand"
)

type UUID [2]uint64

func MakeUUID(low uint64, high uint64) UUID {
	return UUID{low, high}
}

func (u UUID) IsZero() bool {
	return (u[0] | u[1]) == 0
}

func (u UUID) String() string {
	i := uint(0)
	buffer := make([]byte, 37)
	j := 0

	for _, k := range [...]uint{32, 48, 64, 80, 128} {
		for ; i < k; i += 16 {
			x := (u[i/64] >> (i % 64)) & 0xFFFF
			buffer[j] = hexDigits[(x>>4)&0xF]
			buffer[j+1] = hexDigits[x&0xF]
			buffer[j+2] = hexDigits[x>>12]
			buffer[j+3] = hexDigits[(x>>8)&0xF]
			j += 4
		}

		buffer[j] = '-'
		j++
	}

	return string(buffer[:36])
}

func GenerateUUID4() (UUID, error) {
	var buffer [16]byte

	if _, err := crand.Read(buffer[:]); err != nil {
		return UUID{}, err
	}

	uuid := UUID{binary.LittleEndian.Uint64(buffer[:8]), binary.LittleEndian.Uint64(buffer[8:])}
	uuid[0] = (uuid[0] & 0xFF0FFFFFFFFFFFFF) | 0x0040000000000000
	uuid[1] = (uuid[1] & 0xFFFFFFFFFFFFFF3F) | 0x0000000000000080
	return uuid, nil
}

func GenerateUUID4Fast() UUID {
	var buffer [16]byte
	mrand.Read(buffer[:])
	uuid := UUID{binary.LittleEndian.Uint64(buffer[:8]), binary.LittleEndian.Uint64(buffer[8:])}
	uuid[0] = ((uuid[0] ^ mask[0]) & 0xFF0FFFFFFFFFFFFF) | 0x0040000000000000
	uuid[1] = ((uuid[1] ^ mask[1]) & 0xFFFFFFFFFFFFFF3F) | 0x0000000000000080
	return uuid
}

var hexDigits = [...]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}
var mask UUID

func init() {
	var buffer [16]byte

	if _, err := crand.Read(buffer[:]); err != nil {
		panic(err)
	}

	mask = UUID{binary.LittleEndian.Uint64(buffer[:8]), binary.LittleEndian.Uint64(buffer[8:])}
}
