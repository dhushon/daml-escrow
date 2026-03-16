# PLAN.md --- Stablecoin Escrow Platform

## Project Objective

Build a privacy-preserving **stablecoin escrow platform** using **DAML
smart contracts** and **Go-based services** for integration, APIs, and
orchestration.

The system will support:

- Milestone-based escrow
- Mediated dispute resolution
- Oracle-based triggers
- Enterprise-grade privacy using DAML/Canton

------------------------------------------------------------------------

## Development Environment

For local development and testing, we will use the **DAML Sandbox**. It is an in-memory, lightweight DAML ledger that runs locally.

**Rationale:**

- **Simplicity:** Easy to start, stop, and reset.
- **Speed:** Fast for local development and running tests.
- **Compatibility:** Code developed against the sandbox is compatible with other DAML-enabled ledgers, including Canton, ensuring a smooth transition to production environments.
- **No external dependencies:** It does not require a separate database.

For later-stage integration testing that requires persistence, we may introduce a PostgreSQL-backed DAML driver.

------------------------------------------------------------------------

## Phase 0 --- Architecture & Foundation (COMPLETE)

Duration: 2--3 weeks

Goals:

• Finalize system architecture
• Define contract models
• Establish development environment

Deliverables:

- Architecture document (`architecture/system_design.md`)
- Ledger topology design (DAML Sandbox)
- API specification (`architecture/api_specification.yaml`)
- Stablecoin abstraction layer (`internal/ledger/stablecoin.go`)
- DAML contract skeletons
- Initial Go service skeleton

Tasks:

1. Define escrow lifecycle (DONE)
2. Define stablecoin abstraction layer (DONE)
3. Select Canton / ledger environment (DONE)
4. Establish repository structure (DONE)
5. Create CI/CD pipeline (DONE)
6. Define API contracts (DONE)

------------------------------------------------------------------------

## Phase 1 --- Core Escrow Contracts

Duration: 4 weeks

Goals:

Build the base escrow smart contracts.

Deliverables:

- Escrow DAML template
- Settlement template
- Milestone extension model

Tasks:

- Implement escrow template
- Implement settlement contract
- Add milestone choices
- Implement dispute flows
- Write DAML unit tests

Success Criteria:

- Funds lock correctly
- Authorized parties only can release
- Dispute flow returns funds safely

------------------------------------------------------------------------

## Phase 2 --- Backend Platform (Go)

Duration: 6 weeks

Goals:

Create service layer to interact with the ledger.

Services:

• escrow-api
• oracle-service
• payment-gateway
• identity-service

Key responsibilities:

- Contract submission
- Ledger query services
- Event streaming
- Oracle integrations

Tasks:

1. Build REST/gRPC API
2. Implement ledger client
3. Implement escrow creation endpoint
4. Implement milestone approval endpoint
5. Implement dispute endpoint

------------------------------------------------------------------------

## Phase 3 --- Oracle Integrations

Duration: 3 weeks

Goals:

Enable automated triggers.

Examples:

- Shipping confirmation
- Delivery confirmation
- Time-based releases

Architecture:

External System → Oracle Service → DAML Choice Execution

Tasks:

- Oracle adapter interface
- Webhook ingestion
- Signature verification
- Event mapping to contract choices

------------------------------------------------------------------------

## Phase 4 --- Frontend Applications

Duration: 4--6 weeks

Applications:

Buyer Portal
Seller Portal
Mediator Dashboard

Capabilities:

- Escrow creation
- Milestone tracking
- Dispute handling
- Payment confirmation

------------------------------------------------------------------------

## Phase 5 --- Security & Compliance

Duration: 3 weeks

Tasks:

- Contract audit
- Key management review
- Identity integration
- KYC integration hooks

------------------------------------------------------------------------

## Phase 6 --- Production Deployment

Tasks:

- Kubernetes deployment
- Ledger cluster setup
- Observability stack
- Backup and disaster recovery

------------------------------------------------------------------------

## Engineering Milestones

| Milestone | What/Why/Value |
| --------------------- | ------------------------------- |
| M1 | Escrow contract working |
| M2 | API operational |
| M3 | Oracle integration functional |
| M4 | End-to-end escrow flow complete |
| M5 | Production readiness |

------------------------------------------------------------------------

## Risks

| Risk | Mitigation |
| --------------------- | ------------------------------- |
| Ledger complexity | Use managed Canton environments |
| Oracle trust issues | Multi-source validation |
| Stablecoin compliance | Integrate KYC layer |
| Smart contract bugs | Formal testing |

------------------------------------------------------------------------

## Long Term Extensions

- DAO mediation pools
- Marketplace SDK
- Cross-chain settlement
- Tokenized invoice financing
