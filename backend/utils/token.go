package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken menghasilkan string token heksadesimal acak kriptografis
func GenerateToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}
