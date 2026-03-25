# PLAN.md --- Stablecoin Escrow Platform

## Project Objective

Build a privacy-preserving **stablecoin escrow platform** using **DAML
smart contracts** and **Go-based services** for integration, APIs, and
orchestration.

------------------------------------------------------------------------

## Phase 5 --- High-Assurance Lifecycle & Identity (COMPLETE)

Goals: Align contract dynamics with formal escrow standards and establish a production-grade identity bridge.

### Achievements
1. **Robust Escrow Lifecycle:** Refactored DAML and Go to support the standard `DRAFT → FUNDED → ACTIVE → DISPUTED → PROPOSED → SETTLED` sequence per `ESCROW-PROCESS.md`.
2. **Tripartite Authority:** Enforced co-signature requirements for Buyer, Seller, and Issuer (Bank) on terms agreement and settlement ratification.
3. **Bilateral Dispute Resolution:** Implemented a two-stage mediated settlement flow requiring explicit ratification from both disputing parties.
4. **Dynamic Discovery:** Unified Go client to dynamically resolve `packageID` and `partyID` at runtime, ensuring portability across environments.
5. **JIT Identity Bridge:** Established automated mapping between Okta OIDC `sub` claims and unique DAML parties via the User Management API.
6. **Secure Onboarding:** Implemented `EscrowInvitation` flow with token-based "Claim" mechanics to bridge anonymous invitations to authenticated identities.

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
