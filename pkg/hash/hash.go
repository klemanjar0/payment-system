package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// unsafe
const defaultKey = "9f4a2b8c1e7d3f6a0b5e4c2d8f1a9e7b"

type Hasher struct {
	key []byte
}

func New() *Hasher {
	return &Hasher{key: []byte(defaultKey)}
}

func NewWithKey(key string) *Hasher {
	return &Hasher{key: []byte(key)}
}

func (h *Hasher) Hash(input string) string {
	mac := hmac.New(sha256.New, h.key)
	mac.Write([]byte(input))
	return hex.EncodeToString(mac.Sum(nil))
}
