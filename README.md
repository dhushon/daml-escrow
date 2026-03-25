# Stablecoin Escrow Platform (DAML-Based)

## Overview

This project explores the design of a **privacy-preserving, multi‑party
stablecoin escrow platform** implemented using **DAML (Digital Asset
Modeling Language)** and deployed on a **Canton-style distributed ledger
network**.

------------------------------------------------------------------------

## Architecture & Workflow

### System Architecture
The platform is built on a modern, decoupled stack ensuring high performance and cryptographic certainty.

```mermaid
graph TD
    User((User / Admin)) <-->|Port 8080| Astro[Astro Frontend]
    Astro <-->|REST / JSON| Go[Go Backend API]
    Go <-->|Port 8081| JSON[Daml JSON API V2]
    JSON <-->|Port 7575| Canton[Canton Ledger]
    Canton <--> DB[(Postgres DB)]
    Oracle((Oracle Service)) -->|Webhook| Go
```

### Escrow Lifecycle Workflow (Adjudicator Model)
A high-assurance, multi-party flow where stakeholders sign the agreement and an independent Adjudicator (Mediator) facilitates the evidence of completion.

```mermaid
sequenceDiagram
    participant B as Buyer (Stakeholder)
    participant S as Seller (Stakeholder)
    participant M as Mediator (Adjudicator)
    participant L as Daml Ledger
    participant CB as Issuer (Central Bank)

    Note over B,M: 1. Onboarding & Appointment
    B->>L: Create EscrowInvitation (Buyer + Issuer + Mediator)
    L-->>S: Anonymous Preview (Token-based)
    S->>L: Claim Invitation (Verified by Email match)
    L->>L: Create EscrowProposal

    Note over S,L: 2. Agreement Formation
    S->>L: Exercise 'Accept'
    L->>L: Create Active Escrow (Signatories: B, S, CB)

    Note over M,L: 3. Adjudicated Evidence
    M->>L: Approve Milestone (Authoritative Release)
    L->>L: Create EscrowSettlement

    Note over CB,L: 4. Final Settlement
    CB->>L: Settle Payment
    L->>L: Archive Settlement, Funds Transferred
    Note right of S: Seller receives stablecoins
```

------------------------------------------------------------------------

## Key Tools & Helper Apps

### Ledger Identity Sync
To maintain a self-healing, high-assurance architecture, the platform includes a **Sync Helper** that maps logical names to cryptographic identities.

```bash
# Authoritatively DISCOVER and EXPORT current Package and Party IDs
make sync
```
This generates a `ledger-state.json` file which the Go backend automatically loads at startup to synchronize with the latest ledger environment.

------------------------------------------------------------------------

## Getting Started

### Prerequisites
- **Go 1.24+**
- **Java 17 (LTS)**
- **DPM (Daml Package Manager)**
- **Docker & Docker Compose**

### Development Environment

The project provides a unified management console via `Makefile`.

```bash
# View all available strategies
make help

# START the full stack (Ledger + API + Frontend)
make up

# STOP everything
make down
```

### Verification
Verify the full lifecycle functionality using the integration tests:

```bash
# Runs standard unit tests
make test

# Runs full ledger integration suite (requires active ledger)
make integration-test
```

------------------------------------------------------------------------

## Repository Structure

```text
/cmd                    - Entry points for API and Oracle Simulator
/internal/api           - REST handlers and middleware
/internal/ledger        - Modular Daml JSON API V2 client
/internal/services      - Core business logic and metrics orchestration
/contracts/stablecoin-* - Multi-package Daml contracts (Interface, Implementation, Tests)
/frontend               - Astro-based dashboard with Tailwind CSS
/docs                   - API documentation (Swagger)
/.gemini                - Project memory and architectural guardrails
```

------------------------------------------------------------------------

## Product Goals

1. **Enable trust-minimized escrow using stablecoins**
2. **Support milestone-based payments**
3. **Allow dispute mediation**
4. **Provide private contract execution**
5. **Integrate external triggers (oracles, webhooks)**

------------------------------------------------------------------------

## Key Achievements (Phase 5)

- **OIDC Identity Alignment:** Authoritative identity matching using Google Cloud Identity claims.
- **Safe Onboarding:** Token-based invitations with anonymous previews and email-matched claims.
- **Adjudicator Model:** High-assurance lifecycle where Mediator authoritatively backs evidence of completion.
- **Self-Healing Architecture:** Dynamic Package/Party ID discovery with automated local state synchronization.
- **Multi-Tenant Ready:** Thread-safe, context-driven session management for independent user ledgers.
