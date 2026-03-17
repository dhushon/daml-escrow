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

For local development and testing, we use the **Canton 3.4.11 Sandbox** with **DPM**.

**Key Architectural Decisions (SDK 3.x Migration):**
- **Protocol:** Daml JSON Ledger API V2 over HTTP/1.1.
- **Identity:** User-Managed Identities (Stable User IDs mapped to dynamic Canton Party IDs).
- **Tooling:** Mandatory JDK 17 and DPM (Daml Package Manager).

------------------------------------------------------------------------

## Phase 0 --- Architecture & Foundation (COMPLETE)

Duration: 2--3 weeks

Tasks:
1. Define escrow lifecycle (DONE)
2. Define stablecoin abstraction layer (DONE)
3. Select Canton / ledger environment (DONE - Daml 3.4.11)
4. Establish repository structure (DONE)
5. Create CI/CD pipeline (DONE)
6. Define API contracts (DONE)

------------------------------------------------------------------------

## Phase 1 --- Core Escrow Contracts (COMPLETE)

Tasks:
- Implement escrow template (DONE)
- Implement settlement contract (DONE)
- Add milestone choices (DONE)
- Implement dispute flows (DONE)
- Write DAML unit tests (DONE)

------------------------------------------------------------------------

## Phase 2 --- Backend Platform (Go) (IN PROGRESS)

Goals: Create service layer to interact with the ledger.

### Completed Work
1. **REST API:** Swagger UI integrated on port 8080.
2. **JSON V2 Client:** Implemented `JsonLedgerClient` with:
   - Dynamic Party ID resolution via `/v2/parties`.
   - Idempotent bootstrapping via `init.canton`.
   - Recursive V2 JSON response parsing.
   - Ledger offset tracking for consistent reads.
3. **Integration Testing:** Substantial tests in `internal/ledger/ledger_integration_test.go` covering Standard and Refund lifecycles.

### Detailed Next Steps (De-risking & Deep Integration)

**Task 2.7: Multi-Milestone API Expansion**
- Update `internal/ledger/stablecoin.go` to support a `Milestones` slice in the `CreateEscrowRequest`.
- Refactor `internal/ledger/json_client.go` to dynamically build the milestone list in the `CreateCommand` payload.
- Update `internal/api/handlers.go` to accept milestone arrays from the REST client.

**Task 2.8: Mediator "ResolveDispute" Implementation**
- Implement `ResolveDispute` method in `JsonLedgerClient`.
- Add `POST /escrows/{id}/resolve` endpoint to the REST API.
- Update `EscrowService` to handle mediation logic (splitting payouts between Buyer and Seller).

**Task 2.9: Precise Event Extraction**
- Refactor `CreateEscrow` to extract the `contractId` directly from the `events` slice in the `submit-and-wait` response.
- Eliminate the 1-second sleep and "query-after-create" logic to improve performance and thread safety.

**Task 2.10: Stablecoin Settlement Logic**
- Design the hand-off between the Escrow contract and actual token transfers.
- Determine if the Go backend will trigger a secondary `TokenTransfer` contract choice upon milestone approval.

------------------------------------------------------------------------

## Phase 3 --- Oracle Integrations

Tasks:
- Oracle adapter interface
- Webhook ingestion (e.g., Shipping/Delivery confirmations)
- Signature verification
- Event mapping to contract choices (`ApproveMilestone`)

------------------------------------------------------------------------

## Phase 4 --- Frontend Applications

Applications:
- **Buyer Portal:** Escrow creation, milestone approval.
- **Seller Portal:** Progress tracking, payment view.
- **Mediator Dashboard:** Dispute resolution interface.

------------------------------------------------------------------------

## Phase 5 --- Security & Compliance

Tasks:
- JWT Authentication integration (RS256) for the JSON API.
- Contract audit.
- Key management strategy.

------------------------------------------------------------------------

## Engineering Milestones

| Milestone | Status | Description |
| :--- | :--- | :--- |
| M1 | DONE | Escrow contract logic verified in DAML. |
| M2 | DONE | API protocol established (JSON V2). |
| M3 | TODO | Oracle integration functional. |
| M4 | TODO | End-to-end multi-milestone settlement. |
| M5 | TODO | Production readiness (Auth/TLS/Persistence). |
