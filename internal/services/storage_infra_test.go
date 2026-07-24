//go:build integration
package services

import (
	"context"
	"daml-escrow/internal/ledger"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEndToEndStorageMirroring_Infra(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	// 1. Setup Authoritative Storage
	os.Setenv("STORAGE_ENDPOINT", "http://localhost:9000")
	os.Setenv("STORAGE_ACCESS_KEY", "escrow")
	os.Setenv("STORAGE_SECRET_KEY", "escrow-secret")
	bucket := "escrow-bank"
	
	storage, err := NewStorageService(ctx, bucket)
	require.NoError(t, err)

	// 1.5 Pre-provision Authoritative Buckets
	for _, b := range []string{"escrow-bank", "escrow-depositor", "escrow-beneficiary"} {
		_, _ = storage.client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(b),
		})
	}

	// 2. Setup Services
	mockAI := new(MockAIProvider)
	schema, _ := NewSchemaService("../../architecture/schemas")
	ingest := NewIngestService(logger, mockAI, schema, nil, storage)
	
	mockLedger := new(ledger.MockLedgerClient)
	// We use the real storage service inside EscrowService to test mirroring
	escrowSvc := NewEscrowService(logger, mockLedger, nil, nil, "secret", nil, storage, nil)

	// --- STEP 1: Authoritative Ingest (Bank) ---
	t.Run("Bank Ingest and Vaulting", func(t *testing.T) {
		pdfPath := "../../samples/SEC Sample FORM OF ESCROW AGREEMENT.pdf"
		fileData, err := os.ReadFile(pdfPath)
		require.NoError(t, err)

		// Authoritative Mock Setup
		allFileData := [][]byte{fileData}
		mockAI.On("ClassifyContract", mock.Anything, allFileData, "application/pdf").Return("Corporate", nil)
		mockAI.On("ExtractTerms", mock.Anything, allFileData, "application/pdf", "Corporate", mock.Anything).
			Return(`{"amount": 1000.0, "currency": "USD", "beneficiaryEmail": "jimmy@beneficiary.com"}`, nil)
		
		result, err := ingest.IngestContract(ctx, allFileData, "application/pdf")
		require.NoError(t, err)
		assert.NotEmpty(t, result.StorageURI)
		assert.NotEmpty(t, result.ContentHash)
		assert.Contains(t, result.StorageURI, bucket)

		// Verify blob exists in Bank Vault
		downloaded, err := storage.DownloadFromBucket(ctx, bucket, result.StorageURI[len("storage://"+bucket+"/"):])
		assert.NoError(t, err)
		assert.Equal(t, len(fileData), len(downloaded))

		// --- STEP 2: Read-Through Retrieval (Depositor) ---
		t.Run("Depositor Read-Through Mirroring", func(t *testing.T) {
			escrowID := "test-escrow-999"
			userID := "joey@depositor.com"
			depositorBucket := "escrow-depositor"

			// Prepare mock ledger metadata
			meta := ledger.EscrowMetadata{
				AgreementURI: result.StorageURI,
				ContentHash:  result.ContentHash,
			}
			metaJSON, _ := json.Marshal(meta)

			mockLedger.On("GetEscrow", mock.Anything, escrowID, userID).Return(&ledger.EscrowContract{
				ID:          escrowID,
				Depositor:   "Depositor",
				Beneficiary: "Beneficiary",
				Metadata:    string(metaJSON),
			}, nil)

			// TRIGGER READ-THROUGH
			escrow, err := escrowSvc.GetEscrow(ctx, escrowID, userID)
			require.NoError(t, err)

			// VERIFY SIGNED URL
			assert.NotEmpty(t, escrow.AgreementURL)
			assert.Contains(t, escrow.AgreementURL, depositorBucket)
			assert.Contains(t, escrow.AgreementURL, "X-Amz-Signature")

			// VERIFY LOCAL MIRROR (Physical existence in depositor bucket)
			key := result.StorageURI[len("storage://"+bucket+"/"):]
			mirrored, err := storage.DownloadFromBucket(ctx, depositorBucket, key)
			assert.NoError(t, err, "Blob should have been authoritatively mirrored to depositor vault")
			assert.Equal(t, len(fileData), len(mirrored))

			// VERIFY INTEGRITY (Hash check)
			isValid, err := storage.VerifyIntegrity(ctx, depositorBucket, key, result.ContentHash)
			assert.NoError(t, err)
			assert.True(t, isValid, "Mirrored blob hash must match authoritative ledger hash")

			// CLEANUP
			t.Run("Authoritative Cleanup", func(t *testing.T) {
				err1 := storage.Delete(ctx, bucket, key)
				err2 := storage.Delete(ctx, depositorBucket, key)
				assert.NoError(t, err1)
				assert.NoError(t, err2)
			})
		})
	})
}
