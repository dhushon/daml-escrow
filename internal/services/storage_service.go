package services

import (
	"bytes"
	"context"
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
	bucket        string
}

func NewStorageService(ctx context.Context, bucket string) (*StorageService, error) {
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
		// Use static credentials for local development
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
		bucket:        bucket,
	}, nil
}

func (s *StorageService) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	kmsKeyID := os.Getenv("STORAGE_KMS_KEY_ID")
	
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	// High-Assurance: Authoritatively enforce encryption at rest if KMS key is provided
	// Compatible with GCP CMEK (Customer-Managed Encryption Keys)
	if kmsKeyID != "" {
		input.ServerSideEncryption = types.ServerSideEncryptionAwsKms
		input.SSEKMSKeyId = aws.String(kmsKeyID)
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload to storage: %w", err)
	}

	// High-Assurance: Return a logical URI. In production, this would be a signed URL or CloudFront link.
	return fmt.Sprintf("storage://%s/%s", s.bucket, key), nil
}

func (s *StorageService) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	// High-Assurance: Authoritatively restrict access to a time-limited signed URL
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiration))
	
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

func (s *StorageService) Download(ctx context.Context, key string) ([]byte, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download from storage: %w", err)
	}
	defer func() { _ = output.Body.Close() }()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(output.Body)
	return buf.Bytes(), err
}
