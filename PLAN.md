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
5. **Testing Hierarchy:** Established three-tier testing strategy (Unit, Integration, Distributed) with reusable lifecycle logic and deterministic readiness checks.
6. Advanced Analytics Dashboard: Implemented high-assurance **Operational Velocity** visualization with stage latency heatmaps and settlement funnel analytics.
7. **Institutional Multi-Custody:** Implemented full-stack support for **BitGo Enterprise** and **Circle WaaS** via a dynamically pluggable **Stablecoin Factory** architecture.
8. **CIP-0056 Compliance:** Fully refactored the escrow lifecycle to use authoritative cryptographic locking and multi-actor co-signing per the CIP-0056 standard.


### Tasks

**Task 6.1: Multi-Participant Setup (COMPLETE)**

- [x] Distributed Canton topology configuration (Local).
- [x] Distributed Go API Gateway / Multi-Client support.
- [x] Cross-node integration tests (Shared logic for local and distributed).
- [x] Deterministic topology propagation checks in bootstrap.

**Task 6.2: Real Stablecoin Integration (CIP-0056) (COMPLETE)**

- [x] Integrate **CIP-0056** "holding" and "transfer" interfaces into DAML contracts.
- [x] Implement simulation logic for high-assurance token disbursement.
- [x] Research and adapt **OpenZeppelin Canton-Stablecoin** templates for CDP-style vaults.
- [x] Connect with **USDCx** (BitGo/Circle) institutional providers via Dynamic Factory.

**Task 6.3: Analytics & Validation (Noves) (COMPLETE)**

- [x] Integrate **Noves Data & Analytics API** simulation for real-time deposit confirmation.
- [x] Build **LifecycleTracker** process map logic in backend service.
- [x] Create **LifecycleTracker.astro** frontend component for stepwise visualization.
- [x] Build full-cycle metrics dashboard (time-to-complete, bottleneck analysis) using Noves indexed data.

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

## Phase 9 --- High-Assurance Identity & Deep Health (COMPLETE)

Goals: Establish production-grade OIDC bridge, automated infrastructure, and aggregated system diagnostics.

### Achievements

1. **Okta OIDC Bridge:** Implemented strict JWT validation against Okta JWKS. External identity assertions drive Just-In-Time (JIT) ledger provisioning.
2. **Identity-as-Code:** Automated the entire Okta stack (Apps, Servers, Users) using Terraform.
3. **Deep Health Aggregator:** Created a recursive diagnostic engine that aggregates the state of Postgres, Canton, and Oracle sub-systems with latency tracking.
4. **Directory Service:** Added a live counterparty discovery mechanism, allowing users to select authorized participants directly from the ledger.
5. **Production CLI:** Migrated the Go API to a professional Cobra/Viper structure, supporting environment-aware configuration via flags and YAML.

### Tasks

**Task 9.1: Automated OIDC Infrastructure (COMPLETE)**

- [x] Implement Terraform for Okta (Applications, Servers, Policies).
- [x] Configure OIDC Scope-to-Ledger-Permission mapping.
- [x] Automated Test Persona provisioning script (`setup_test_users.sh`).

**Task 9.2: Home Realm Discovery & Discovery (COMPLETE)**

- [x] Create `/auth/discovery` endpoint in Go API.
- [x] Implement email-domain based routing to correct IdP in Astro.
- [x] Implement Directory Service (`/api/v1/identities`) for on-ledger discovery.

**Task 9.3: System Integrity & Diagnostics (COMPLETE)**

- [x] Implement aggregated health checks for all dependencies.
- [x] Build live health cockpit in frontend footer.
- [x] Standardize 10-retry hardened template readiness for ledger connections.

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
