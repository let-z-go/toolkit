package uuid

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type UUID [16]byte

func (self UUID) IsZero() bool {
	return (self[0x0] | self[0x1] | self[0x2] | self[0x3] |
		self[0x4] | self[0x5] | self[0x6] | self[0x7] |
		self[0x8] | self[0x9] | self[0xA] | self[0xB] |
		self[0xC] | self[0xD] | self[0xE] | self[0xF]) == 0
}

func (self UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", self[:4], self[4:6], self[6:8], self[8:10], self[10:])
}

func (self UUID) Base64() string {
	return base64.RawURLEncoding.EncodeToString(self[:])
}

func GenerateUUID4() (UUID, error) {
	var uuid UUID

	if _, e := rand.Read(uuid[:]); e != nil {
		return uuid, e
	}

	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] & 0x3F) | 0x80
	return uuid, nil
}

func UUIDFromBytes(bytes []byte) UUID {
	var uuid UUID
	copy(uuid[:], bytes)
	return uuid
}
