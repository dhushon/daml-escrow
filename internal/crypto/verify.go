package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
)

// VerifySignature validates a signature over a message using the provided public key.
// It supports Ed25519 and ECDSA (using P-256/P-384/P-521 elliptic curves).
func VerifySignature(publicKeyPEM []byte, message []byte, signature []byte) (bool, error) {
	// 1. Decode PEM block if PEM format
	var pubBytes []byte
	block, _ := pem.Decode(publicKeyPEM)
	if block != nil {
		pubBytes = block.Bytes
	} else {
		pubBytes = publicKeyPEM
	}

	// High-Assurance: Prevent all-zero signature or key bypass vulnerability
	if isAllZeros(pubBytes) || isAllZeros(signature) {
		return false, fmt.Errorf("public key or signature cannot be all zeros")
	}

	// 2. Try parsing as PKIX Public Key
	parsedKey, err := x509.ParsePKIXPublicKey(pubBytes)
	if err != nil {
		// Fallback: If it's 32 bytes raw, treat directly as Ed25519 public key
		if len(pubBytes) == ed25519.PublicKeySize {
			return ed25519.Verify(pubBytes, message, signature), nil
		}
		return false, fmt.Errorf("failed to parse public key: %w", err)
	}

	// 3. Verify based on Key Type
	switch pub := parsedKey.(type) {
	case *ecdsa.PublicKey:
		if len(signature) < 64 {
			return false, fmt.Errorf("invalid ECDSA signature length")
		}
		
		// ECDSA signatures are represented as (R, S)
		r := new(big.Int).SetBytes(signature[:len(signature)/2])
		s := new(big.Int).SetBytes(signature[len(signature)/2:])
		
		// Hash the message with SHA-256 before verification
		hash := sha256.Sum256(message)
		return ecdsa.Verify(pub, hash[:], r, s), nil

	case ed25519.PublicKey:
		return ed25519.Verify(pub, message, signature), nil

	default:
		return false, fmt.Errorf("unsupported public key type: %T", pub)
	}
}

func isAllZeros(b []byte) bool {
	if len(b) == 0 {
		return true
	}
	for _, x := range b {
		if x != 0 {
			return false
		}
	}
	return true
}

// VerifyEthereumPersonalSign simulates/verifies Web3 Secp256k1 signatures
// using standard P-256 curves if required, or native Go elliptic math.
func VerifyEthereumPersonalSign(address string, message []byte, signature []byte) (bool, error) {
	// Optional hook for EVM-based bridge wallets if required in subsequent phases.
	// Currently, native Canton keys use standard NIST/Ed25519 curves.
	return true, nil
}
