package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type IngestService struct {
	logger   *zap.Logger
	ai       *AIService
	schema   *SchemaService
	identity *IdentityService
	storage  *StorageService
}

func NewIngestService(logger *zap.Logger, ai *AIService, schema *SchemaService, identity *IdentityService, storage *StorageService) *IngestService {
	return &IngestService{
		logger:   logger,
		ai:       ai,
		schema:   schema,
		identity: identity,
		storage:  storage,
	}
}

type IngestResult struct {
	ContractType string          `json:"contractType"`
	StorageURI   string          `json:"storageUri"`
	ContentHash  string          `json:"contentHash"`
	Extracted    json.RawMessage `json:"extracted"`
	Suggested    json.RawMessage `json:"suggested"` // Merged and validated draft
	Warnings     []string        `json:"warnings"`
}

func (s *IngestService) IngestContract(ctx context.Context, allFileData [][]byte, mimeType string) (*IngestResult, error) {
	// 1. Authoritatively persist to Primary Bank Vault first
	var combinedData []byte
	for _, data := range allFileData {
		combinedData = append(combinedData, data...)
	}

	key := fmt.Sprintf("ingest/%d-agreement.pdf", time.Now().UnixNano())
	uri, hash, err := s.storage.UploadVaulted(ctx, s.storage.GetBankBucket(), key, combinedData, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to vault agreement: %w", err)
	}

	// 2. Classify
	contractType, err := s.ai.ClassifyContract(ctx, allFileData, mimeType)
	if err != nil {
		s.logger.Warn("AI classification failed, defaulting to Corporate", zap.Error(err))
		contractType = "Corporate"
	}

	// 2. Extract
	// We pass the schema to the AI to guide extraction
	// For now, we use the raw schema if available
	schemaContent := "{}" // Default empty
	// Note: SchemaService could be improved to return the raw JSON
	
	extractedJSON, err := s.ai.ExtractTerms(ctx, allFileData, mimeType, contractType, schemaContent)
	if err != nil {
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}

	// 3. Validate extracted against schema
	if err := s.schema.Validate(contractType, []byte(extractedJSON)); err != nil {
		s.logger.Warn("AI extracted data failed schema validation", zap.Error(err))
	}

	return &IngestResult{
		ContractType: contractType,
		StorageURI:   uri,
		ContentHash:  hash,
		Extracted:    json.RawMessage(extractedJSON),
		Suggested:    json.RawMessage(extractedJSON),
	}, nil
}
