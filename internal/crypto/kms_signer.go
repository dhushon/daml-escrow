package crypto

import (
	"context"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
)

// CloudKMSSigner implements HighAssuranceSigner using Google Cloud KMS (HSM-backed).
type CloudKMSSigner struct {
	client *kms.KeyManagementClient
	keyID  string
}

func NewCloudKMSSigner(ctx context.Context, keyID string) (*CloudKMSSigner, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}
	return &CloudKMSSigner{
		client: client,
		keyID:  keyID,
	}, nil
}

func (s *CloudKMSSigner) Close() error {
	return s.client.Close()
}

func (s *CloudKMSSigner) Sign(ctx context.Context, digest []byte) ([]byte, error) {
	req := &kmspb.AsymmetricSignRequest{
		Name: s.keyID,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: digest,
			},
		},
	}

	result, err := s.client.AsymmetricSign(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("KMS signing failed: %w", err)
	}

	return result.Signature, nil
}

func (s *CloudKMSSigner) PublicKey(ctx context.Context) ([]byte, error) {
	req := &kmspb.GetPublicKeyRequest{
		Name: s.keyID,
	}

	result, err := s.client.GetPublicKey(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public key from KMS: %w", err)
	}

	return []byte(result.Pem), nil
}
