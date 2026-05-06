package crypto

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
)

// HighAssuranceSigner defines the authoritative interface for signing operations.
type HighAssuranceSigner interface {
	Sign(ctx context.Context, digest []byte) ([]byte, error)
	PublicKey(ctx context.Context) ([]byte, error)
}

// LocalSigner implements a software-based signer for CI and standalone development.
// It generates an ephemeral key for logic verification, ensuring ZERO cloud dependency.
type LocalSigner struct {
	privateKey *ecdsa.PrivateKey
}

func NewLocalSigner() (*LocalSigner, error) {
	// Use authoritative elliptic curve for ECDSA generation
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate local ephemeral key: %w", err)
	}
	return &LocalSigner{privateKey: key}, nil
}

func (s *LocalSigner) Sign(ctx context.Context, digest []byte) ([]byte, error) {
	r, sig, err := ecdsa.Sign(rand.Reader, s.privateKey, digest)
	if err != nil {
		return nil, err
	}
	return append(r.Bytes(), sig.Bytes()...), nil
}

func (s *LocalSigner) PublicKey(ctx context.Context) ([]byte, error) {
	// SA1019: Use modern crypto/ecdh for public key serialization
	ecdhPub, err := s.privateKey.PublicKey.ECDH()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ECDH public key: %w", err)
	}
	return ecdhPub.Bytes(), nil
}
