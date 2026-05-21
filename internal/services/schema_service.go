package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
)

type SchemaService struct {
	schemas map[string]*gojsonschema.Schema
}

func NewSchemaService(schemaDir string) (*SchemaService, error) {
	s := &SchemaService{
		schemas: make(map[string]*gojsonschema.Schema),
	}

	files, err := os.ReadDir(schemaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		typeName := file.Name()[:len(file.Name())-len(".json")]
		schemaPath := filepath.Join(schemaDir, file.Name())
		
		// High-Assurance: Absolute path required for gojsonschema file loader
		absPath, _ := filepath.Abs(schemaPath)
		loader := gojsonschema.NewReferenceLoader("file://" + absPath)
		
		schema, err := gojsonschema.NewSchema(loader)
		if err != nil {
			return nil, fmt.Errorf("failed to load schema %s: %w", typeName, err)
		}

		s.schemas[typeName] = schema
	}

	return s, nil
}

func (s *SchemaService) Validate(contractType string, data []byte) error {
	schema, ok := s.schemas[contractType]
	if !ok {
		// If no specific schema, allow anything (for now) or enforce default
		return nil 
	}

	loader := gojsonschema.NewBytesLoader(data)
	result, err := schema.Validate(loader)
	if err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if !result.Valid() {
		var errs string
		for _, desc := range result.Errors() {
			errs += fmt.Sprintf("- %s\n", desc)
		}
		return fmt.Errorf("data does not conform to %s schema:\n%s", contractType, errs)
	}

	return nil
}

func (s *SchemaService) GetSupportedTypes() []string {
	types := make([]string, 0, len(s.schemas))
	for t := range s.schemas {
		types = append(types, t)
	}
	return types
}
