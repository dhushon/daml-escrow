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

## Phase 6 --- Scaling & Distribution (IN PROGRESS)

Goals: Move to multi-participant topology and real stablecoin integration.

### Achievements

1. **Multi-Participant Infrastructure:** Created `contracts/DevNet/devnet.conf` and robust `devnet_init.canton` for distributed topology with deterministic propagation checks.
2. **Local Distributed Environment:** Added `docker-compose.distributed.yml` to enable local testing of separate Bank, Buyer, and Seller nodes.
3. **Frontend Lifecycle Alignment:** Refactored `api.ts` and `EscrowCard.astro` to support the full high-assurance state machine (`DRAFT` → `FUNDED` → `ACTIVE` → `DISPUTED` → `PROPOSED` → `SETTLED`).
4. **Distributed Gateway:** Implemented `MultiLedgerClient` in Go for role-based node routing and resilient identity resolution across clusters.
5. **Testing Hierarchy:** Established three-tier testing strategy: Unit (Fast/Mocks), Integration (Single-node), and Distributed (Multi-node).

### Tasks

**Task 6.1: Multi-Participant Setup (COMPLETE)**

- [x] Distributed Canton topology configuration (Local).
- [x] Distributed Go API Gateway / Multi-Client support.
- [x] Cross-node integration tests (Shared logic for local and distributed).
- [x] Deterministic topology propagation checks in bootstrap.

**Task 6.2: Real Stablecoin Integration (CIP-0056)**

- [ ] Integrate **CIP-0056** "holding" and "transfer" interfaces into DAML contracts.
- [ ] Research and adapt **OpenZeppelin Canton-Stablecoin** templates for CDP-style vaults.
- [ ] Connect with **USDCx** (BitGo/Circle) mock/testnet assets for collateral pledging.
- [ ] Implement **Splice Validator API** integration for external signing and automated agent workflows.

**Task 6.3: Analytics & Validation (Noves)**

- [ ] Integrate **Noves Data & Analytics API** for real-time deposit confirmation.
- [ ] Implement "Lock Validation" logic: Confirm funds are on-ledger before activating the escrow.
- [ ] Build transaction history and wallet metrics dashboard using Noves indexed data.

------------------------------------------------------------------------

## Phase 8 --- Distributed Sovereignty & Multi-Service Deployment

Goals: Transition to isolated service-per-node architecture.

### Tasks

**Task 8.1: Service Containerization**

- [ ] Refactor `main.go` to support single-node mode via environment variables.
- [ ] Create `Dockerfile` for Go API and Frontend.
- [ ] Implement health checks per service role.

**Task 8.2: Distributed API Routing**

- [ ] Implement API Gateway (Nginx/Envoy) with JWT-based routing logic.
- [ ] Configure cross-service telemetry for the distributed topology.

**Task 8.3: Compliance Validation**

- [ ] Perform "Node Isolation" security audit.
- [ ] Verify "Right to be Forgotten" patterns across distributed storage nodes.

------------------------------------------------------------------------

## Phase 9 --- Advanced Identity Infrastructure (SAML + OIDC + HRD)

Goals: Finalize full "Terraformed" Identity Provider (IdP) and enterprise onboarding.

### Tasks

**Task 9.1: Multi-Protocol IdP (Terraforming)**

- [ ] Implement Terraform for Google Cloud Identity Platform (GCIP).
- [ ] Configure OIDC (Google/Okta) and SAML (Enterprise) provider configurations.
- [ ] Enable Multi-Tenancy for specific corporate domains.

**Task 9.2: Home Realm Discovery (HRD) Implementation**

- [ ] Create `/auth/discovery` endpoint in Go API.
- [ ] Implement email-domain based routing to SAML vs OIDC in Astro.

**Task 9.3: Enterprise Onboarding Workflow**

- [ ] Build "Claim with SSO" flow in `onboard.astro`.
- [ ] Implement JIT Provisioning for SAML attributes (Mapping SAML Groups to Daml Roles).
- [ ] Automated KYC/Compliance webhook integration (Mock).

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
