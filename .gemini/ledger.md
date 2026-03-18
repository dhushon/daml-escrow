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
- **Synchronous Returns:** Use `/v2/commands/submit-and-wait-for-transaction` to get the `contractId` immediately in the response `events` slice.
- **State Discovery:** `/v2/state/active-contracts` for querying the current Active Contract Set (ACS).

---

## 2. Development Rules & Patterns

### Rule 1: Numeric Precision (Decimal)
Daml `Decimal` types (Map to Go `float64`) MUST be serialized as strings with exactly **10 decimal places**.
- **Go Pattern:** `fmt.Sprintf("%.10f", value)`
- **Failure Mode:** Providing fewer decimal places results in `COMMAND_PREPROCESSING_FAILED`.

### Rule 2: JSON Key Casing
- **Transaction Commands:** Must use **PascalCase** (e.g., `CreateCommand`, `ExerciseCommand`).
- **Nesting:** Commands MUST be nested under a top-level `commands: { ... }` key for `/v2/commands/submit-and-wait-for-transaction`.

### Rule 3: ACS Query Visibility (Lessons Learned)
- **Failure:** Querying `/v2/state/active-contracts` with `activeAtOffset: 0` often returns an empty list immediately after a transaction, even if the transaction is confirmed.
- **Retry Strategy:** If a fetch fails, implement a short retry loop (e.g., 5 attempts with 500ms delay).
- **Offset Omission:** To ensure the query hits the absolute ledger head, OMIT the `activeAtOffset` field entirely. This forces the JSON API to return the most recent state.

---

## 3. Integration Failures & Resolutions

| Issue | Attempted Fix | Outcome | Final Solution |
| :--- | :--- | :--- | :--- |
| `escrow not found` after create | `activeAtOffset: 0` | FAILED (Stale data) | Omit `activeAtOffset` to query head. |
| `WRONGLY_TYPED_CONTRACT` | Used old ID after Choice | FAILED | Extract NEW `contractId` from response events. |
| `Missing required field actAs` | Top-level fields | FAILED | Nest under `commands: { ... }` for V2. |

---

## 4. Expectations for Development Work

1. **Idempotent Bootstrapping:** The `init.canton` script must be runnable multiple times.
2. **Event Extraction:** ALWAYS prefer extracting IDs from the transaction response rather than re-querying the ledger.
3. **Milestone Consistency:** When creating multi-milestone escrows, ensure `totalAmount` exactly matches the sum of milestones to satisfy the DAML `ensure` clause.

---

## 5. Transition to Production

| Feature | Development (Current) | Production Requirement |
| :--- | :--- | :--- |
| **Storage** | Memory | Persistent (Postgres/Oracle) |
| **Security** | No Auth / Sandbox Defaults | JWT Authentication (RS256/ES256) |
| **Synchronizer** | `local-sync` (In-process) | External Sequencer (BFT/Database) |
| **TLS** | Disabled (Plaintext) | Mandatory Mutual TLS (mTLS) |

---

## 6. Verification Checklist

- [ ] `make sandbox` starts the node.
- [ ] `make ledger-setup` establishes users and topology.
- [ ] `make integration-test` passes (Create -> Fetch -> Multi-Milestone -> Dispute -> Resolve).
