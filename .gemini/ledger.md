# Daml Ledger Integration Guide (SDK 3.4.11)

This document defines the architectural patterns, development rules, and production path for the Canton ledger integration.

## 1. Core Integration Strategy

### User-Managed Identities
To avoid the instability of Canton's long, random Party IDs (e.g., `Buyer::1220...`), we use **Daml User Management**.
- **Strategy:** The Go backend interacts with the ledger using human-readable **User IDs** (`Buyer`, `Seller`, `CentralBank`).
- **Mapping:** These User IDs are mapped to the actual Canton parties during the bootstrap phase (`init.canton`).
- **Authorization:** Permissions (e.g., `actAs`) are granted to the User, allowing the backend to submit commands without managing private keys directly in development.

### JSON Ledger API V2
We utilize the modern V2 API, which aligns closely with gRPC but remains accessible over HTTP.
- **Protocol:** HTTP/1.1 (Standard for the JSON API proxy).
- **Communication Pattern:** `submit-and-wait` for synchronous command execution.
- **State Discovery:** `/v2/state/active-contracts` for querying the current Active Contract Set (ACS).

---

## 2. Development Rules & Patterns

### Rule 1: Numeric Precision (Decimal)
Daml `Decimal` types (Map to Go `float64`) MUST be serialized as strings with exactly **10 decimal places**.
- **Go Pattern:** `fmt.Sprintf("%.10f", value)`
- **Failure Mode:** Providing fewer decimal places or a raw number may result in a `COMMAND_PREPROCESSING_FAILED` error.

### Rule 2: JSON Key Casing
- **Commands:** Must use **PascalCase** (e.g., `CreateCommand`, `ExerciseCommand`).
- **Fields:** Must use **camelCase** (e.g., `templateId`, `createArguments`, `choiceArgument`).
- **userId:** Every command submission MUST include the `userId` field if auth is disabled or JWT claims are missing.

### Rule 3: Dynamic Party Resolution
The Go client MUST resolve User IDs to full Canton Party IDs at initialization.
- **Reason:** Contract arguments inside the ledger (e.g., the `buyer` field in a template) require the full `Party::Hex` string, even if the submission uses a User ID.
- **Pattern:** Query `/v2/parties` and build a prefix-based lookup map.

---

## 4. Expectations for Development Work

1. **Idempotent Bootstrapping:** The `init.canton` script must be runnable multiple times against a persistent or in-memory ledger without failing. Use `ignore` or check-before-create logic.
2. **Topology Readiness:** Topology changes (party enablement) take time to propagate across the synchronizer. Always include a small delay (~2-5s) after topology transactions in setup scripts.
3. **Template References:** Use the format `Module:Template` for template IDs to maintain package independence, or the full `PackageID:Module:Template` for absolute precision.

---

## 5. Transition to Production

When moving from local Sandbox to a Production/Testnet environment, the following changes are required:

| Feature | Development (Current) | Production Requirement |
| :--- | :--- | :--- |
| **Storage** | Memory | Persistent (Postgres/Oracle) |
| **Security** | No Auth / Sandbox Defaults | JWT Authentication (RS256/ES256) |
| **Synchronizer** | `local-sync` (In-process) | External Sequencer (BFT/Database) |
| **TLS** | Disabled (Plaintext) | Mandatory Mutual TLS (mTLS) |
| **Identity** | Automated Bootstrap | Formal Onboarding / DID Integration |
| **API Proxy** | Local Process | Highly Available (Load Balanced) |

---

## 6. Verification Checklist

- [ ] `make sandbox` starts the node.
- [ ] `make ledger-setup` establishes users and topology.
- [ ] `curl http://localhost:7575/v2/parties` returns all expected IDs.
- [ ] `make integration-test` passes (Create -> Fetch -> Exercise).
