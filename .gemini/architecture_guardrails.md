# Architecture Guardrails

These guardrails define the **core architectural principles** for the escrow platform.

## 1. Contract-First Design

All business logic involving funds MUST live in DAML contracts. Backend services orchestrate, validate inputs, and relay transactions. They must NEVER implement financial logic.

## 2. Ledger as Source of Truth

The distributed ledger is the **single source of truth**. Applications must treat the ledger as authoritative. No off-ledger balance tracking is allowed.

## 3. Privacy by Default

Escrow contracts must restrict visibility to: **Buyer, Seller, Mediator, Issuer**. No escrow information should leak to unrelated parties.

## 4. Identity Bridge & JIT Provisioning (Phase 5+)

- **Identity Source:** Google Cloud Identity (GCIP) is the authoritative IdP.
- **Mapping:** External Subject IDs (`sub`) MUST be mapped to Daml User IDs via a deterministic sanitization process.
- **JIT Allocation:** New users must be provisioned on-ledger at first-login via the Daml User Management API, allocating a unique cryptographic Party ID.
- **Principle of Least Privilege:** Users must only be granted `actAs` rights for their own primary party.

## 5. Stateless Services

All Go services must be stateless. Persistent state must reside in the ledger or authorized external stores (e.g. `user_config` database).

## 6. Namespaced Configuration

Module-specific configuration variables MUST be prefixed with the module name (e.g., `AUTH_ISSUER`, `LEDGER_HOST`) to prevent global collision and improve discoverability.

## 7. Metadata & Oracle Guardrails

- **Schema Maturity:** Every business domain must provide a versioned JSON Schema in `/architecture/schemas`.
- **Minimalist Ledger:** Only data required for settlement or audit-linkage should be persisted to the ledger.
- **Privacy Redaction:** Sensitive operational data MUST use the `exclusions` pattern to prevent leakage to the immutable record.
- **Oracle Trust:** Webhooks must be authenticated using HMAC-SHA256 signatures verified against `ORACLE_WEBHOOK_SECRET`.

## 8. Self-Healing Lifecycle (Phase 5+)

- **Dynamic Bindings:** All contract references (Package IDs) and participant identifiers (Party IDs) MUST be resolved at runtime during the startup discovery phase.
- **Resilience:** Systems must survive ledger resets and contract re-builds by automatically re-synchronizing their identity maps and version tags from the authoritative ledger metadata.

## 10. Testing & Validation Hierarchy (Phase 6)

- **MANDATORY:** All features MUST include tests in one of the three established tiers:
    1.  **Unit Tier (`go test ./...`):** Logic-only, using mocks (e.g., `MockLedgerClient`). MUST be fast and require zero network/Docker access.
    2.  **Integration Tier (`-tags=integration`):** Single-node Sandbox validation. Ensures smart contract logic is correct against a live API.
    3.  **Distributed Tier (`-tags=distributed`):** Multi-node Canton topology validation. Focuses on cross-node authorization, routing, and topology propagation.
- **Shared Logic:** Reuse core lifecycle validation logic (e.g., `shared_test.go`) across Integration and Distributed tiers to ensure behavioral consistency.

## 11. Distributed Identity & Routing

- **Zero-Trust Routing:** The `MultiLedgerClient` MUST act as a smart gateway, routing commands to the specific participant node hosting the user's party.
- **Identity Probing:** Systems MUST NOT assume a user exists on a single node. User lookups MUST be performed across the cluster cluster until found or all nodes exhausted.
- **Deterministic Readiness:** Infrastructure scripts (Canton bootstrap) MUST NOT use temporal waits (`sleep`). They MUST use deterministic polling of the topology state until the required conditions (e.g., cross-node party visibility) are met.
