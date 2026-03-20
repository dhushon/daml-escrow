# Architecture Guardrails

These guardrails define the **core architectural principles** for the
escrow platform.

## 1. Contract-First Design

All business logic involving funds MUST live in DAML contracts.

Backend services:

- orchestrate
- validate inputs
- relay transactions

They must NEVER implement financial logic.

## 2. Ledger as Source of Truth

The distributed ledger is the **single source of truth**.

Applications must treat the ledger as authoritative.

No off-ledger balance tracking is allowed.

## 3. Privacy by Default

Escrow contracts must restrict visibility to:

- buyer
- seller
- mediator
- issuer

No escrow information should leak to unrelated parties.

## 4. Stateless Services

All Go services must be stateless.

State must reside in:

- the ledger
- durable event streams

## 5. Event-Driven Architecture

Services should subscribe to ledger events rather than polling.

Preferred technologies:

- Kafka
- NATS
- Cloud PubSub

## 6. Deterministic Escrow Flows

Escrow state transitions must always follow deterministic rules:

Created → Locked → Delivered → Released/Refunded

## 7. Metadata & Oracle Guardrails

- **Schema Maturity:** Every business domain must provide a versioned JSON Schema in `/architecture/schemas`.
- **Minimalist Ledger:** Only data required for settlement or audit-linkage should be persisted to the ledger.
- **Privacy Redaction:** Sensitive operational data (detailed locations, PII, operator codes) MUST use the `exclusions` ("don't event") pattern to prevent leakage to the immutable record.
- **Oracle Trust:** Webhooks must be authenticated using HMAC or asymmetric signatures verified against a pre-shared secret or public key.

## 8. Modular Integration Pattern

Large ledger clients or complex integration layers MUST be refactored into modular files to prevent monolithic corruption and improve maintainability.

- **Specialization:** Separate logic into files by concern: `base`, `parser`, `parties`, and domain-specific entities (e.g., `escrows`, `settlements`).
- **Shared State:** Shared structs and constants must reside in the `base` file or a `generated` package.
- **AI Tooling Safety:** Modularization ensures that automated surgical edits remain high-signal and low-risk.
