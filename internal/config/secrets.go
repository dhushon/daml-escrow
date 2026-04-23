package config

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// SecretResolver provides high-assurance access to cloud-based secrets.
type SecretResolver struct {
	client    *secretmanager.Client
	projectID string
}

func NewSecretResolver(ctx context.Context, projectID string) (*SecretResolver, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}
	return &SecretResolver{
		client:    client,
		projectID: projectID,
	}, nil
}

func (r *SecretResolver) Close() error {
	return r.client.Close()
}

// GetSecret retrieves the latest version of a secret from GCP Secret Manager.
func (r *SecretResolver) GetSecret(ctx context.Context, secretName string) (string, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", r.projectID, secretName)
	
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := r.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", secretName, err)
	}

	return string(result.Payload.Data), nil
}
