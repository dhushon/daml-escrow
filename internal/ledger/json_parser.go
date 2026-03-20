package ledger

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// parseNDJSON decodes either a single JSON array or newline-delimited JSON objects.
func parseNDJSON(body []byte) ([]map[string]interface{}, error) {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil, nil
	}

	var results []map[string]interface{}

	// Check if it's a JSON array
	if body[0] == '[' {
		if err := json.Unmarshal(body, &results); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON array: %w", err)
		}
		return results, nil
	}

	// Otherwise treat as NDJSON
	lines := bytes.Split(body, []byte("\n"))
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var item map[string]interface{}
		if err := json.Unmarshal(line, &item); err != nil {
			return nil, fmt.Errorf("NDJSON line parse error: %w", err)
		}
		results = append(results, item)
	}
	return results, nil
}
