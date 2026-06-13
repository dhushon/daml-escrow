#!/bin/bash
set -e

# Resolve paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOT_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"

SWAGGER_JSON="$ROOT_DIR/docs/swagger.json"
OPENAPI3_JSON="$ROOT_DIR/docs/openapi3.json"
OUTPUT_TS="$ROOT_DIR/frontend/src/types/api.d.ts"

echo "Generating TypeScript types from OpenAPI spec..."
if [ ! -f "$SWAGGER_JSON" ]; then
    echo "Error: Swagger documentation (swagger.json) not found at $SWAGGER_JSON"
    echo "Please run 'make swagger-gen' in the root first."
    exit 1
fi

# Create directory
mkdir -p "$(dirname "$OUTPUT_TS")"

# Convert Swagger 2.0 to OpenAPI 3.0 temporarily
echo "Converting Swagger 2.0 to OpenAPI 3.0..."
npx -y swagger2openapi "$SWAGGER_JSON" -o "$OPENAPI3_JSON"

# Run generator
npx openapi-typescript "$OPENAPI3_JSON" -o "$OUTPUT_TS"

# Clean up temp file
rm -f "$OPENAPI3_JSON"

echo "TypeScript declarations generated successfully at: frontend/src/types/api.d.ts"
