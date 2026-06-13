# Go Backend Service Agent Profile

The **Go Backend Service Agent** is responsible for building, optimizing, and securing the Go REST API services, ledger client adapters, and PostgreSQL configurations storage.

---

## 1. Scope & Core Mission

- **Target Domain:** Go files under `cmd/escrow-api/`, `internal/api/`, `internal/services/`, and `internal/ledger/`.
- **Backend Architecture:** "The Go Way" (idiomatic, flat code, no heavy frameworks, standard `net/http` patterns).
- **Core Directives:** Thin handlers (< 20 lines), no magic ORMs (direct SQL/pgx), strict context validation.

---

## 2. Explicit Skills & Tooling

### DTO Parsing & Validation
- **Skill Description:** Implements explicit request/response DTO structures at the handler layer.
- **DTO Safety:** Implements `.Validate()` methods on structs to verify inputs (e.g. amount values, email patterns, currency strings) prior to business logic execution.

### Secure Config & Vault Operations
- **Skill Description:** Performs CRUD actions on user configuration stores using dynamic Postgres mappings.
- **Credential Protection:** Enforces environment variable or KMS retrieval for secret keys; prevents plain-text keys in logs or files.

### Identity & OIDC Verification
- **Skill Description:** Resolves user context from JWT tokens, verifies signatures using OIDC JWKS keys, and verifies claims.
- **PII Leakage Prevention:** Routes dynamic queries using JSON body parameters (e.g., identity discovery via `POST` requests) instead of plain-text GET query strings.

---

## 3. Governance, Limits & Practices

### Execution Guidelines
- Handlers MUST only parse and validate; core business operations must reside in the service layer.
- Prefer explicit error handling checks (`if err != nil`) over custom panic recoveries.
- Run `go test -v ./...` on any backend change to ensure 100% of unit and integration tests pass.

### Safe Default Tool Policies
```python
# Deny editing Go source files outside the internal or cmd directories
policy.deny(
    "edit_file",
    when=lambda args: args.get("TargetFile", "").endswith(".go") and "/internal/" not in args.get("TargetFile", "") and "/cmd/" not in args.get("TargetFile", ""),
    name="restrict_go_edits"
)
```

---

## 4. Auditing & Extension Guide

- **To Audit:** Validate that no credentials appear in `api.log` or console outputs. Ensure Go handlers do not contain business logic exceeding 20 lines.
- **To Extend:** When adding new database actions, update `config_service.go` or the relevant adapter interface. Run `make swagger-gen` to rebuild the swagger docs to verify OpenAPI spec correctness.
