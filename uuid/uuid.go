package uuid

import (
	srand "crypto/rand"
	"encoding/base64"
	"math/rand"
	"time"
)

type UUID [16]byte

func (self UUID) IsZero() bool {
	return (self[0x0] | self[0x1] | self[0x2] | self[0x3] |
		self[0x4] | self[0x5] | self[0x6] | self[0x7] |
		self[0x8] | self[0x9] | self[0xA] | self[0xB] |
		self[0xC] | self[0xD] | self[0xE] | self[0xF]) == 0
}

func (self UUID) String() string {
	i := 0
	rawString := make([]byte, 36)
	j := 0

	for _, k := range [...]int{4, 6, 8, 10} {
		for ; i < k; i++ {
			x := self[i]
			rawString[j] = hexDigits[x>>4]
			rawString[j+1] = hexDigits[x&0xF]
			j += 2
		}

		rawString[j] = '-'
		j++
	}

	for ; i < 16; i++ {
		x := self[i]
		rawString[j] = hexDigits[x>>4]
		rawString[j+1] = hexDigits[x&0xF]
		j += 2
	}

	return string(rawString)
}

func (self UUID) Base64() string {
	return base64.RawURLEncoding.EncodeToString(self[:])
}

func GenerateUUID4() (UUID, error) {
	var uuid UUID

	if _, e := srand.Read(uuid[:]); e != nil {
		return uuid, e
	}

	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] & 0x3F) | 0x80
	return uuid, nil
}

func GenerateUUID4Fast() UUID {
	var uuid UUID
	rand_.Read(uuid[:])

	for i := range uuid {
		uuid[i] ^= mask[i]
	}

	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] & 0x3F) | 0x80
	return uuid
}

func UUIDFromBytes(bytes []byte) UUID {
	var uuid UUID
	copy(uuid[:], bytes)
	return uuid
}

var hexDigits = [...]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}
var rand_ *rand.Rand
var mask [16]byte

func init() {
	rand_ = rand.New(rand.NewSource(time.Now().UnixNano()))

	if _, e := srand.Read(mask[:]); e != nil {
		panic(e)
	}
}
