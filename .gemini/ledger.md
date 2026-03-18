# Daml Ledger Integration Guide (SDK 3.4.11)

This document defines the architectural patterns, development rules, and production path for the Canton ledger integration.

## 1. Core Integration Strategy

### User-Managed Identities
To avoid the instability of Canton's long, random Party IDs (e.g., `Buyer::1220...`), we use **Daml User Management**.
- **Strategy:** The Go backend interacts with the ledger using human-readable **User IDs** (`Buyer`, `Seller`, `CentralBank`, `EscrowMediator`).
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

### Rule 3: ACS Query Visibility
- **Challenge:** Querying `/v2/state/active-contracts` immediately after a transaction can return stale data due to asynchronous indexing.
- **Solution:** 
  1. Implement a retry loop (5 attempts, 500ms delay) in `GetEscrow`.
  2. ALWAYS use the latest ledger offset retrieved from `/v2/state/ledger-end` to ensure consistent reads.

### Rule 4: Package ID Management
- **Challenge:** Any change to DAML code results in a new Package ID, which breaks existing backend references.
- **Resolution:** After any DAML change, use `unzip -l contracts/.daml/dist/*.dar` to identify the new main DALF hash and update `PackageID` in the Go client.

---

## 3. Integration Failures & Resolutions

| Issue | Attempted Fix | Outcome | Final Solution |
| :--- | :--- | :--- | :--- |
| `escrow not found` after create | `activeAtOffset: 0` | FAILED (Stale data) | Use `/v2/state/ledger-end` offset. |
| `WRONGLY_TYPED_CONTRACT` | Used old ID after Choice | FAILED | Extract NEW `contractId` from response events. |
| `Invalid template` | Rebuilt DAML | FAILED (Missing choice) | Add `Settle` choice and update `PackageID`. |

---

## 4. Settlement Lifecycle

The escrow uses a two-phase settlement pattern:
1.  **Approval:** `buyer` approving a milestone creates an `EscrowSettlement` contract.
2.  **Execution:** `issuer` (Central Bank) exercises the `Settle` choice on the `EscrowSettlement` to finalize the payment and release stablecoins.

---

## 5. Verification Checklist

- [ ] `make restart-ledger` cleans, builds, and starts the environment.
- [ ] `make ledger-setup` establishes users and topology.
- [ ] `make integration-test` passes (Create -> Fetch -> Multi-Milestone -> Dispute -> Resolve -> Settle).
