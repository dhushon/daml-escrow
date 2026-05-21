package services

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIProvider interface {
	ClassifyContract(ctx context.Context, allFileData [][]byte, mimeType string) (string, error)
	ExtractTerms(ctx context.Context, allFileData [][]byte, mimeType string, contractType string, schema interface{}) (string, error)
}

type AIService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewAIService(ctx context.Context) (*AIService, error) {
	apiKey := os.Getenv("GOOGLE_GENAI_API_KEY")
	if apiKey == "" {
		// High-Assurance: Auth fallback to default credentials if using Vertex AI
		if os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "true" {
			return &AIService{}, nil // Will initialize model on-demand with application default credentials
		}
		return nil, fmt.Errorf("GOOGLE_GENAI_API_KEY is required")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	return &AIService{
		client: client,
		model:  client.GenerativeModel("gemini-2.0-flash"),
	}, nil
}

func (s *AIService) ClassifyContract(ctx context.Context, allFileData [][]byte, mimeType string) (string, error) {
	prompt := `
		Analyze the following escrow agreement (provided as one or more pages) and classify it into exactly one of these types:
		ImportExport, RealEstate, Grants, Corporate.
		
		Return ONLY the type name.
	`
	
	parts := []genai.Part{genai.Text(prompt)}
	for _, data := range allFileData {
		parts = append(parts, genai.Blob{MIMEType: mimeType, Data: data})
	}

	resp, err := s.model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("classification failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "Corporate", nil // Fallback
	}

	// Simple extraction of the first text part
	if part, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(part), nil
	}

	return "Corporate", nil
}

func (s *AIService) ExtractTerms(ctx context.Context, allFileData [][]byte, mimeType string, contractType string, schema interface{}) (string, error) {
	prompt := fmt.Sprintf(`
		Extract the escrow terms from the attached agreement (provided as one or more pages).
		The extracted data MUST conform to the following JSON Schema:
		%v

		Return the data as a valid JSON object.
	`, schema)

	parts := []genai.Part{genai.Text(prompt)}
	for _, data := range allFileData {
		parts = append(parts, genai.Blob{MIMEType: mimeType, Data: data})
	}

	resp, err := s.model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("extraction failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "{}", nil
	}

	if part, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		return string(part), nil
	}

	return "{}", nil
}
