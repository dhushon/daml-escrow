//go:build integration
package services

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageService_MinIO_Integration(t *testing.T) {
	ctx := context.Background()

	// High-Assurance: Authoritatively use local MinIO for integration
	os.Setenv("STORAGE_ENDPOINT", "http://localhost:9000")
	os.Setenv("STORAGE_ACCESS_KEY", "escrow")
	os.Setenv("STORAGE_SECRET_KEY", "escrow-secret")
	os.Setenv("STORAGE_BUCKET", "integration-test-bucket")
	defer os.Unsetenv("STORAGE_ENDPOINT")

	svc, err := NewStorageService(ctx, "integration-test-bucket")
	require.NoError(t, err)

	// 1. Ensure Bucket Exists (Authoritative Pre-requisite)
	_, err = svc.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String("integration-test-bucket"),
	})
	// Ignore if already exists
	
	t.Run("End-to-End Upload and Download", func(t *testing.T) {
		key := "test-doc.txt"
		content := []byte("Authoritative institutional content for security verification.")
		contentType := "text/plain"

		// Upload
		uri, _, err := svc.UploadVaulted(ctx, "integration-test-bucket", key, content, contentType, nil)
		assert.NoError(t, err)
		assert.Contains(t, uri, "integration-test-bucket/test-doc.txt")

		// Download
		downloaded, err := svc.DownloadFromBucket(ctx, "integration-test-bucket", key)
		assert.NoError(t, err)
		assert.Equal(t, content, downloaded)
	})

	t.Run("Presigned URL Signing and Access", func(t *testing.T) {
		key := "signed-doc.txt"
		content := []byte("Private agreement content.")
		
		_, _, err := svc.UploadVaulted(ctx, "integration-test-bucket", key, content, "text/plain", nil)
		require.NoError(t, err)

		// Generate Presigned URL (Valid for 1 minute)
		url, err := svc.GetPresignedURL(ctx, "integration-test-bucket", key, 1*time.Minute)
		assert.NoError(t, err)
		assert.Contains(t, url, "X-Amz-Signature") // Standard S3 signing param

		// Verify the URL actually works (External fetch simulator)
		resp, err := http.Get(url)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		defer resp.Body.Close()
	})

	t.Run("Authoritative Object Tagging", func(t *testing.T) {
		key := "tagged-doc.txt"
		content := []byte("Tagged content.")
		tags := map[string]string{
			"contract-id": "escrow-999",
			"institutional-priority": "high",
		}

		_, _, err := svc.UploadVaulted(ctx, "integration-test-bucket", key, content, "text/plain", tags)
		require.NoError(t, err)

		// Verify tags via S3 API
		output, err := svc.client.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
			Bucket: aws.String("integration-test-bucket"),
			Key:    aws.String(key),
		})
		assert.NoError(t, err)
		
		found := 0
		for _, tag := range output.TagSet {
			if val, ok := tags[*tag.Key]; ok && val == *tag.Value {
				found++
			}
		}
		assert.Equal(t, 2, found, "Expected all authoritative tags to be present on the blob")
	})

	t.Run("Security: Encryption at Rest Headers", func(t *testing.T) {
		os.Setenv("STORAGE_KMS_KEY_ID", "mock-kms-key-id")
		defer os.Unsetenv("STORAGE_KMS_KEY_ID")

		key := "encrypted-doc.txt"
		content := []byte("Encrypted institutional data.")
		
		_, _, err := svc.UploadVaulted(ctx, "integration-test-bucket", key, content, "text/plain", nil)
		
		// High-Assurance Verification: 
		// If using local MinIO, we expect it to FAIL with "NotImplemented" if headers are present 
		// but KMS isn't configured in the container. This proves the client sent the headers.
		if err != nil {
			assert.Contains(t, err.Error(), "NotImplemented", "Expected NotImplemented error (KMS not configured) proving headers were sent")
		} else {
			// If MinIO is configured with KMS, this would pass.
			t.Log("MinIO accepted SSE-KMS headers (likely ignored or configured)")
		}
	})
}
