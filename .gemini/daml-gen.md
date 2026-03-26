# DAML & Go Integration Rules

## Integration Strategy

- **Primary Client:** Use `JsonLedgerClient` (JSON API V2) for all dynamic binding requirements. This ensures compatibility across ledger resets without requiring Go code recompilation.
- **Dynamic Discovery:** Every interaction MUST rely on the `Discover(ctx)` phase to resolve current Package IDs and Party IDs.

## Critical Version Compatibility

- **Ledger API Version:** The backend services are built for **Ledger API v2** (JSON and gRPC).
- **SDK Requirement:** **Daml SDK 3.0 or higher** is MANDATORY.
- **Daml-LF Target:** Use `2.1` for SDK 3.4.x compatibility.

## Java Environment

- **Requirement:** **JDK 17** (LTS).
- **Rationale:** Required for Daml 3.0+ CLI tools.

## Tooling

- **SDK 3.0+ Management:** Use **`dpm`** (Digital Asset Package Manager) where available, otherwise `daml` CLI with 3.x configuration.

## Contract Design Patterns (Formal Escrow)

- **State Machine:** Use specific templates for each lifecycle state (`EscrowProposal`, `EscrowContract`, `DisputeRecord`, etc.) rather than a single template with an enum. This enforces strict authority guards.
- **Authority:** `Issuer` MUST be a signatory on all state templates to ensure control over disbursement.
- **Interface Implementation:** Use the SDK 3.x interface pattern. Define stable data structures in the interface package and logic in the implementation package.

## Ledger Connection

- **Host:** Prefer `127.0.0.1` over `localhost`.
- **JSON API V2:** Target the `/v2/commands/submit-and-wait-for-transaction` and `/v2/state/active-contracts` endpoints.
