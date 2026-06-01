package crypto

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifySignature_Ed25519(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	message := []byte("Authoritative challenge nonce")
	signature := ed25519.Sign(privKey, message)

	t.Run("Positive: Valid Signature", func(t *testing.T) {
		valid, err := VerifySignature(pubKey, message, signature)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("Negative: Tampered Message", func(t *testing.T) {
		valid, err := VerifySignature(pubKey, []byte("Tampered message"), signature)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Negative: Mismatched Key", func(t *testing.T) {
		wrongPub, _, _ := ed25519.GenerateKey(rand.Reader)
		valid, err := VerifySignature(wrongPub, message, signature)
		assert.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Negative: Corrupted Signature", func(t *testing.T) {
		badSig := make([]byte, len(signature))
		copy(badSig, signature)
		badSig[0] ^= 0xFF // corrupt a bit
		valid, err := VerifySignature(pubKey, message, badSig)
		assert.NoError(t, err)
		assert.False(t, valid)
	})
}

func TestVerifySignature_ECDSA_P256(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)

	message := []byte("ECDSA challenge nonce")
	hash := sha256.Sum256(message)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	assert.NoError(t, err)

	// Encode signature as R || S (concatenated)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	// Ensure padding to curve size (32 bytes for P-256)
	signature := make([]byte, 64)
	copy(signature[32-len(rBytes):32], rBytes)
	copy(signature[64-len(sBytes):64], sBytes)

	// Marshal public key to PKIX DER bytes
	pubDer, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	assert.NoError(t, err)

	t.Run("Positive: Valid PKIX Public Key", func(t *testing.T) {
		valid, err := VerifySignature(pubDer, message, signature)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("Positive: Valid PEM Encoded Public Key", func(t *testing.T) {
		pubPem := pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubDer,
		})
		valid, err := VerifySignature(pubPem, message, signature)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("Negative: Short Signature Length", func(t *testing.T) {
		_, err := VerifySignature(pubDer, message, []byte("short"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ECDSA signature length")
	})

	t.Run("Negative: Tampered Message", func(t *testing.T) {
		valid, err := VerifySignature(pubDer, []byte("Tampered message"), signature)
		assert.NoError(t, err)
		assert.False(t, valid)
	})
}

func TestVerifySignature_EdgeCases(t *testing.T) {
	t.Run("Negative: Invalid Public Key Bytes", func(t *testing.T) {
		_, err := VerifySignature([]byte("completely invalid public key"), []byte("msg"), []byte("sig"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse public key")
	})

	t.Run("Negative: Hardened Zeros Check Fallback", func(t *testing.T) {
		// An invalid key that is exactly 32 bytes of zeros should trigger our zero-check and return an error
		badKey := make([]byte, ed25519.PublicKeySize)
		valid, err := VerifySignature(badKey, []byte("msg"), make([]byte, ed25519.SignatureSize))
		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "cannot be all zeros")
	})
}
