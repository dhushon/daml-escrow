# Go Engineering Guardrails

## 1. Language Philosophy

Go services should be small, explicit, and deterministic. Avoid unnecessary abstractions or heavy ORMs.

## 2. Project Layout

Target a standard Go structure:

- `/cmd`: Entry points (e.g. `escrow-api`)
- `/internal`: Private library code (ledger, services, api)
- `/pkg`: Shared public logic (logging)
- `/config`: YAML/Env configuration

## 3. Dependency Rules

- **Minimize Third-Party:** Prefer the standard library.
- **Agnostic Auth:** Use standard JWT libraries (`golang-jwt`) instead of vendor-specific SDKs to ensure cross-platform portability.
- **Preferred Stack:** `chi` (routing), `zap` (logging), `sql/pgx` (database).

## 4. Context-Aware Security (Phase 5+)

- **Identity Propagation:** Authenticated subject IDs MUST be stored in the request context using type-safe keys (e.g. `AuthSubKey`).
- **Scope Enforcement:** Every handler MUST use a helper (e.g. `RequireScope`) to verify permissions before executing business logic.

## 5. Configuration & CLI Strategy

- **CLI Framework:** Use **Cobra** for all service entry points to provide a professional, discoverable CLI (e.g., `serve`, `migrate`, `sync` commands).
- **Configuration Management:** Use **Viper** to bind environment variables and flags to configuration.
- **Structured Config:** Consolidated all configuration into a single, nested `Config` struct. NEVER use global variables or separate `os.Getenv` calls in the service layer.
- **Validation:** Configuration must be loaded into strongly typed structs at startup and validated before use.

## 6. Error Handling

Errors must be explicit and include wrapped context:
`return fmt.Errorf("ledger submission failed: %w", err)`

## 7. API Handlers & Validation

- **Data Transfer Objects (DTOs):** Always define specific Request and Response structs (DTOs) for API endpoints rather than using the core ledger boundary structs directly.
- **Input Validation:** Implement `.Validate() error` on DTOs to catch malformed payloads, negative numbers, or invalid enums *before* they enter the service layer.

## 8. Testing & Quality

- **Unit Tests:** Mandatory for all service logic using interfaces/mocks.
- **Integration Tests:** Required for ledger-facing code using the `integration` build tag.
- **Validation:** `go build ./...` and `go test ./...` must pass before any push.
