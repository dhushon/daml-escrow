package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type StorageService struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bankBucket    string
}

func NewStorageService(ctx context.Context, bankBucket string) (*StorageService, error) {
	var cfg aws.Config
	var err error

	endpoint := os.Getenv("STORAGE_ENDPOINT")
	accessKey := os.Getenv("STORAGE_ACCESS_KEY")
	secretKey := os.Getenv("STORAGE_SECRET_KEY")
	region := os.Getenv("STORAGE_REGION")
	if region == "" {
		region = "us-east-1"
	}

	if endpoint != "" {
		// High-Assurance Local Path: MinIO / S3-compatible
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
	} else {
		// High-Assurance Production Path: GCS S3-Interoperability or Native S3
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load storage config: %w", err)
	}

	// High-Assurance: Authoritatively configure the S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = true // Required for MinIO and GCS S3 Interop
	})

	return &StorageService{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bankBucket:    bankBucket,
	}, nil
}

// UploadVaulted stores a blob in a specific bucket and returns its SHA-256 hash.
func (s *StorageService) UploadVaulted(ctx context.Context, bucket, key string, data []byte, contentType string, tags map[string]string) (string, string, error) {
	kmsKeyID := os.Getenv("STORAGE_KMS_KEY_ID")
	
	// Authoritatively calculate SHA-256 hash for provenance
	hash := sha256.Sum256(data)
	contentHash := hex.EncodeToString(hash[:])

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	if kmsKeyID != "" {
		input.ServerSideEncryption = types.ServerSideEncryptionAwsKms
		input.SSEKMSKeyId = aws.String(kmsKeyID)
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload to vault %s: %w", bucket, err)
	}

	// High-Assurance: Apply tags immediately after upload
	if len(tags) > 0 {
		_ = s.TagObject(ctx, bucket, key, tags)
	}

	return fmt.Sprintf("storage://%s/%s", bucket, key), contentHash, nil
}

// TagObject authoritatively applies contract identities and referential metadata to a blob.
func (s *StorageService) TagObject(ctx context.Context, bucket, key string, tags map[string]string) error {
	var s3Tags []types.Tag
	for k, v := range tags {
		s3Tags = append(s3Tags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	_, err := s.client.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(key),
		Tagging: &types.Tagging{TagSet: s3Tags},
	})
	return err
}

// ReadThroughMirror authoritatively implements the pull-on-demand pattern.
func (s *StorageService) ReadThroughMirror(ctx context.Context, targetBucket, key string, tags map[string]string) ([]byte, error) {
	// 1. Try to fetch from target bucket first
	data, err := s.DownloadFromBucket(ctx, targetBucket, key)
	if err == nil {
		return data, nil
	}

	// 2. If missing, fetch from authoritative Bank Vault
	bankData, err := s.DownloadFromBucket(ctx, s.bankBucket, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from primary vault: %w", err)
	}

	// 3. Lazily mirror to target bucket with tags
	_, _, err = s.UploadVaulted(ctx, targetBucket, key, bankData, "application/pdf", tags)
	if err != nil {
		// Non-blocking warning: serving from bank copy even if local caching fails
		fmt.Printf("warning: failed to mirror blob to %s: %v\n", targetBucket, err)
	}

	return bankData, nil
}

func (s *StorageService) DownloadFromBucket(ctx context.Context, bucket, key string) ([]byte, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = output.Body.Close() }()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(output.Body)
	return buf.Bytes(), err
}

func (s *StorageService) Delete(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *StorageService) GetPresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// GetBankBucket returns the authoritative primary bucket name.
func (s *StorageService) GetBankBucket() string {
	return s.bankBucket
}

// VerifyIntegrity authoritatively checks a local blob's hash against the ledger record.
func (s *StorageService) VerifyIntegrity(ctx context.Context, bucket, key, expectedHash string) (bool, error) {
	data, err := s.DownloadFromBucket(ctx, bucket, key)
	if err != nil {
		return false, err
	}

	hash := sha256.Sum256(data)
	actualHash := hex.EncodeToString(hash[:])
	
	return actualHash == expectedHash, nil
}
