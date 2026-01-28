package api

import (
	"crypto/rand"
	"encoding/hex"
)

func newTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
