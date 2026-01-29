package api

import (
	"crypto/rand"
	"encoding/hex"
)

// newTraceID generates a unique 16-byte hex string for request tracing
func newTraceID() string {
	b := make([]byte, 16)
	// Using cryptographically secure random number generator
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
