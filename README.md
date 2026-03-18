# Stablecoin Escrow Platform (DAML-Based)

## Overview

This project explores the design of a **privacy-preserving, multi‑party
stablecoin escrow platform** implemented using **DAML (Digital Asset
Modeling Language)** and deployed on a **Canton-style distributed ledger
network**.

------------------------------------------------------------------------

## Getting Started

### Prerequisites
- **Go 1.24+**
- **Java 17 (LTS)**
- **DPM (Daml Package Manager)**
- **Docker & Docker Compose**

### Development Environment (Sandbox)

The project supports two main ways to run the ledger environment:

#### 1. Local Sandbox (Recommended for Development)
This mode runs the Canton ledger locally on your host machine while using a Docker container for Postgres persistence.

```bash
# Bring up Postgres, build contracts, start ledger, and setup topology/users
make sandbox-up

# Run the Go API locally
make run
```

#### 2. Full Docker Stack
This mode runs the entire platform (Postgres, Ledger, and Go API) inside Docker containers.

```bash
# Deploy the full persistent stack
make docker-up

# Verify service health
docker-compose ps
```

### Verification
Once the ledger is initialized (either locally or via Docker), verify the full lifecycle functionality using the integration tests:

```bash
# Runs standard unit tests
make test

# Runs full ledger integration suite (requires active ledger)
make integration-test
```

------------------------------------------------------------------------

## Repository Structure

```text
/cmd/escrow-api         - Entry point for the Go API service
/internal/api           - REST handlers and routing
/internal/ledger        - Daml JSON API V2 client
/internal/services      - Escrow orchestration logic
/contracts/Sandbox      - Stable single-node ledger configurations
/contracts/DevNet       - Standard multi-node configurations (UNTESTED)
/contracts/daml         - Daml smart contract templates
/scripts                - Setup and utility scripts
/docs                   - API documentation (Swagger)
```

------------------------------------------------------------------------

## Development Modes

### Sandbox Mode (Stable)
Uses a single-node Canton process. The participant node uses **Postgres** for persistence, while internal infrastructure (Sequencer/Mediator) uses in-memory storage to simplify local development.

### DevNet Mode (UNTESTED)
Located in `/contracts/DevNet`, this configuration mirrors a production Canton network with separate participant and domain nodes. This mode is currently **untested** and intended for future distributed deployments.

------------------------------------------------------------------------

## Product Goals

1. **Enable trust-minimized escrow using stablecoins**
2. **Support milestone-based payments**
3. **Allow dispute mediation**
4. **Provide private contract execution**
5. **Integrate external triggers (oracles, webhooks)**

------------------------------------------------------------------------

## Key Achievements (Phase 2)

- **Daml 3.4.11 / LF 2.1:** Modernized contracts for the latest Canton standards.
- **JSON API V2:** Full integration with the dynamic party and user management system.
- **Persistent Storage:** Integrated Postgres for ledger state preservation.
- **Verified Lifecycle:** 100% test coverage for creation, milestones, disputes, refunds, and settlements.
