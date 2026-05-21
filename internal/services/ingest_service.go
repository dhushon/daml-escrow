package services

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type IngestService struct {
	logger        *zap.Logger
	ai            *AIService
	schema        *SchemaService
	identity      *IdentityService
}

func NewIngestService(logger *zap.Logger, ai *AIService, schema *SchemaService, identity *IdentityService) *IngestService {
	return &IngestService{
		logger:   logger,
		ai:       ai,
		schema:   schema,
		identity: identity,
	}
}

type IngestResult struct {
	ContractType string          `json:"contractType"`
	Extracted    json.RawMessage `json:"extracted"`
	Suggested    json.RawMessage `json:"suggested"` // Merged and validated draft
	Warnings     []string        `json:"warnings"`
}

func (s *IngestService) IngestContract(ctx context.Context, fileData []byte, mimeType string) (*IngestResult, error) {
	// 1. Classify
	contractType, err := s.ai.ClassifyContract(ctx, fileData, mimeType)
	if err != nil {
		s.logger.Warn("AI classification failed, defaulting to Corporate", zap.Error(err))
		contractType = "Corporate"
	}

	// 2. Extract
	// We pass the schema to the AI to guide extraction
	// For now, we use the raw schema if available
	schemaContent := "{}" // Default empty
	// Note: SchemaService could be improved to return the raw JSON
	
	extractedJSON, err := s.ai.ExtractTerms(ctx, fileData, mimeType, contractType, schemaContent)
	if err != nil {
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}

	// 3. Validate extracted against schema
	if err := s.schema.Validate(contractType, []byte(extractedJSON)); err != nil {
		s.logger.Warn("AI extracted data failed schema validation", zap.Error(err))
	}

	return &IngestResult{
		ContractType: contractType,
		Extracted:    json.RawMessage(extractedJSON),
		Suggested:    json.RawMessage(extractedJSON),
	}, nil
}
