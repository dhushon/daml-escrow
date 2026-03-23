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

## Phase 4 --- Frontend & Finalization (COMPLETE)

Goals: Build user-facing dashboard and finalize production readiness.

### Achievements
1. **Astro Dashboard:** Multi-page dashboard for Escrows, Settlements, and Platform Metrics.
2. **End-to-End Lifecycle UX:** Implemented "Composition Wizard" for drafting agreements and a two-party "Propose-Accept" sign-off flow.
3. **Admin & Oversight Mode:** Role-based views (Mediator oversight vs Buyer/Seller action) with persistent cookie-based state.
4. **Advanced Observability:** Real-time system performance (Latency, Uptime, CPU) and ledger health (TPS, Success Rate) visualizations.
5. **Branded UI:** Professional, dark-mode compatible UI using DataCloud LNF standards.
6. **Simulated Ecosystem:** Mocked stablecoin wallet registry and fund transfer visualizations.

------------------------------------------------------------------------

## Phase 5 --- Identity & Onboarding (IN PROGRESS)

Goals: Establish a production-grade identity bridge between Okta and Daml.

### Tasks
**Task 5.1: Okta OIDC Integration**
- Integrate Okta authentication into Astro (Client-side) and Go (Bearer token verification).

**Task 5.2: JIT Party Provisioning**
- Build `IdentityService` to automatically map Okta `sub` IDs to unique Daml Party IDs.
- Implement automated User creation via Daml User Management API.

**Task 5.3: Invitation Lifecycle**
- Implement `EscrowInvitation` Daml template for inviting external emails to a contract.
- Secure token-based "Claim" flow to link anonymous invitations to authenticated Okta users.

**Task 5.4: Hyperlink UX**
- Secure signed-URL generation for email invitations.
- Onboarding landing page for "Unclaimed" participants.

------------------------------------------------------------------------

## Phase 6 --- Scaling & Distribution (NEXT)

Goals: Move to multi-participant topology and real stablecoin integration.

### Tasks
**Task 6.1: Multi-Participant Setup**
- Distributed Canton topology with separate participant nodes for Buyer, Seller, and Bank.

**Task 6.2: Real Stablecoin Integration**
- Integrate with a real stablecoin contract (e.g. Daml Finance or ERC-20 bridge).

------------------------------------------------------------------------

## Phase 7 --- Production Hardening

Goals: Final security and observability sweep.

### Tasks
**Task 7.1: TLS & Key Management**
- TLS 1.3+ configuration for all endpoints.
- HSM/KMS integration for ledger and API signing keys.

**Task 7.2: Full Observability**
- OpenTelemetry (OTEL) integration for cross-service tracing.
- Prometheus/Grafana metrics export.
