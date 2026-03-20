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

## Phase 3 --- Oracle Integrations (COMPLETE)

Goals: Automate milestone approvals via external triggers.

### Achievements
1. **Webhook Ingestion:** Secured endpoint with HMAC-SHA256 signature verification.
2. **Automated Approval Logic:** Validates contract state and milestone indexing before executing automated approvals.
3. **Modular Ledger Architecture:** Refactored Go JSON client into specialized modules for scalability and maintainability.
4. **Daml 3.x Compatibility:** Fully aligned with SDK 3.4.x / Canton authorization and serialization requirements.
5. **Oracle Simulation:** CLI utility for testing end-to-end automation flows.

------------------------------------------------------------------------

## Phase 4 --- Frontend & Finalization (NEXT)

Goals: Build user-facing dashboard and finalize production readiness.

### Tasks
**Task 4.1: Astro Dashboard**
- Create overview of active escrows for Buyers and Sellers.
- Implement "Raise Dispute" and "Approve Milestone" buttons.

**Task 4.2: Settlement View**
- List pending settlements for Central Bank payout.

**Task 4.3: Real-time Updates**
- Integrate WebSockets or polling for ledger state changes.
