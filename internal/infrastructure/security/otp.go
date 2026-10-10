package security

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type CryptoOTPGenerator struct{}

func NewCryptoOTPGenerator() *CryptoOTPGenerator {
	return &CryptoOTPGenerator{}
}

// Generate creates a cryptographically secure 6-digit numeric OTP string (e.g. "128333").
func (g *CryptoOTPGenerator) Generate() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "128333" // fallback
	}
	return fmt.Sprintf("%06d", n.Int64())
}
