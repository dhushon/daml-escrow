package api

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestWalletAuth_ChallengeResponse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	db, smock, err := sqlmock.New()
	require := assert.New(t)
	require.NoError(err)
	defer func() { _ = db.Close() }()

	// 1. Setup Services & Mocks
	configContent := `
providers:
  wallet.devlocal:
    type: OIDC
`
	tmpFile, err := os.CreateTemp("", "idp*.yaml")
	require.NoError(err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.Write([]byte(configContent))
	_ = tmpFile.Close()

	configSvc := services.NewMockConfigService(db)
	mockLedger := new(ledger.MockLedgerClient)
	idSvc, err := services.NewIdentityService(tmpFile.Name(), db)
	require.NoError(err)

	escrowSvc := services.NewEscrowService(logger, mockLedger, nil, nil, "secret", nil, nil, nil)
	h := NewHandler(logger, escrowSvc, nil, configSvc, nil, idSvc, nil, nil, nil)

	// Generate a valid Ed25519 keypair for cryptographic signing
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(err)

	t.Run("GET /auth/nonce - Success", func(t *testing.T) {
		// Mock the insert of the generated nonce into the DB
		smock.ExpectExec("INSERT INTO auth_nonces").
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		req, _ := http.NewRequest("GET", "/api/v1/auth/nonce", nil)
		rr := httptest.NewRecorder()

		h.GetNonce(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		
		var resp map[string]string
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(err)
		assert.Contains(t, resp["nonce"], "Sign this challenge to authenticate:")
		require.NoError(smock.ExpectationsWereMet())
	})

	t.Run("POST /auth/wallet/verify - Valid Signature", func(t *testing.T) {
		challenge := "Sign this challenge to authenticate: test-nonce-123"
		
		// Sign the challenge using the private key
		signature := ed25519.Sign(privKey, []byte(challenge))

		// Encode signature and public key to hex
		sigHex := hex.EncodeToString(signature)
		pubHex := hex.EncodeToString(pubKey)
		partyID := "Depositor::test"

		// 1. Mock the nonce validation (Delete and consume)
		smock.ExpectQuery("DELETE FROM auth_nonces").
			WithArgs(challenge).
			WillReturnRows(sqlmock.NewRows([]string{"nonce"}).AddRow(challenge))

		// 2. Mock identity resolution (GetOrCreateIdentity expects DB lookup)
		oktaSub := "wallet:" + partyID
		email := "depositor@wallet.devlocal"
		
		mockLedger.On("GetIdentity", mock.Anything, oktaSub).Return(&ledger.UserIdentity{
			OktaSub:     oktaSub,
			DamlUserID:  "u-depositor-test",
			DamlPartyID: partyID,
			Email:       email,
			Role:        "Depositor",
		}, nil)

		// Create request body
		verifyReq := VerifyWalletRequest{
			Nonce:       challenge,
			Signature:   sigHex,
			PublicKey:   pubHex,
			DamlPartyId: partyID,
		}
		body, _ := json.Marshal(verifyReq)

		req, _ := http.NewRequest("POST", "/api/v1/auth/wallet/verify", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.VerifyWallet(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(err)
		assert.NotEmpty(t, resp["token"])

		// Cryptographically verify the returned Platform JWT Session Token
		tokenStr := resp["token"].(string)
		parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("platform-jwt-signing-secret-key-32-bytes!"), nil
		})
		require.NoError(err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		assert.True(t, ok)
		assert.Equal(t, oktaSub, claims["sub"])
		assert.Equal(t, email, claims["email"])
		assert.Equal(t, "wallet", claims["auth_method"])
		assert.Equal(t, "daml-escrow-platform", claims["iss"])

		require.NoError(smock.ExpectationsWereMet())
	})

	t.Run("POST /auth/wallet/verify - Replayed/Invalid Nonce", func(t *testing.T) {
		challenge := "Sign this challenge to authenticate: replayed-nonce"
		verifyReq := VerifyWalletRequest{
			Nonce:       challenge,
			Signature:   "sig",
			PublicKey:   "pub",
			DamlPartyId: "Depositor::test",
		}
		body, _ := json.Marshal(verifyReq)

		// Mock nonce verify to fail (no rows deleted / already consumed)
		smock.ExpectQuery("DELETE FROM auth_nonces").
			WithArgs(challenge).
			WillReturnError(sql.ErrNoRows)

		req, _ := http.NewRequest("POST", "/api/v1/auth/wallet/verify", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.VerifyWallet(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid or expired nonce")
	})

	t.Run("POST /auth/wallet/verify - Cryptographic Failure", func(t *testing.T) {
		challenge := "Sign this challenge to authenticate: valid-nonce"
		
		// Sign with a different/wrong key to simulate signature failure
		_, wrongPriv, _ := ed25519.GenerateKey(rand.Reader)
		signature := ed25519.Sign(wrongPriv, []byte(challenge))

		sigHex := hex.EncodeToString(signature)
		pubHex := hex.EncodeToString(pubKey) // Genuine public key, but mismatched signature

		smock.ExpectQuery("DELETE FROM auth_nonces").
			WithArgs(challenge).
			WillReturnRows(sqlmock.NewRows([]string{"nonce"}).AddRow(challenge))

		verifyReq := VerifyWalletRequest{
			Nonce:       challenge,
			Signature:   sigHex,
			PublicKey:   pubHex,
			DamlPartyId: "Depositor::test",
		}
		body, _ := json.Marshal(verifyReq)

		req, _ := http.NewRequest("POST", "/api/v1/auth/wallet/verify", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		h.VerifyWallet(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid signature")
	})
}
