# PLAN.md --- Stablecoin Escrow Platform

## Project Objective

Build a privacy-preserving **stablecoin escrow platform** using **DAML
smart contracts** and **Go-based services** for integration, APIs, and
orchestration.

------------------------------------------------------------------------

## Phase 2 --- Backend Platform (Go) (COMPLETE)

Goals: Create service layer to interact with the ledger.

### Achievements
1. **REST API:** Swagger UI integrated.
2. **JSON V2 Client:** Robust integration with dynamic parties and event extraction.
3. **Milestone & Dispute Logic:** Full lifecycle implemented and tested.
4. **Stablecoin Settlement:** Payout finalization by Central Bank.
5. **Ledger Stability:** Established persistent Sandbox (Postgres) and Docker Compose stack with automated topology/user setup.
6. **Validation:** 100% test pass on backend integration suite.

------------------------------------------------------------------------

## Phase 3 --- Oracle Integrations (CURRENT)

Goals: Automate milestone approvals via external triggers.

### Tasks
**Task 3.1: Webhook Ingestion Service**
- Create endpoint to receive proof-of-delivery or milestone completion events.
- Implement signature verification for trusted oracle providers.

**Task 3.2: Automated Approval Logic**
- Map webhook payloads to active Escrow contract IDs.
- Authorize and execute `ApproveMilestone` choices on behalf of the Oracle/Buyer.

**Task 3.3: External API Integration**
- Mock external shipping/service APIs to trigger webhooks.

------------------------------------------------------------------------

## Phase 4 --- Frontend & Finalization (NEXT)
...
