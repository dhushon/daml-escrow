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

## Phase 6 --- Scaling & Distribution (COMPLETE)

Goals: Move to multi-participant topology and real stablecoin integration.

### Achievements

1. **Multi-Participant Infrastructure:** Created `contracts/DevNet/devnet.conf` and robust `devnet_init.canton` for distributed topology with deterministic propagation checks.
2. **Local Distributed Environment:** Added `docker-compose.distributed.yml` to enable local testing of separate Bank, Buyer, and Seller nodes.
3. **Frontend Lifecycle Alignment:** Refactored `api.ts` and `EscrowCard.astro` to support the full high-assurance state machine (`DRAFT` → `FUNDED` → `ACTIVE` → `DISPUTED` → `PROPOSED` → `SETTLED`).
4. **Distributed Gateway:** Implemented `MultiLedgerClient` in Go for role-based node routing and resilient identity resolution across clusters.
5. **Testing Hierarchy:** Established three-tier testing strategy (Unit, Integration, Distributed) with reusable lifecycle logic and deterministic readiness checks.
6. **Advanced Analytics Dashboard:** Implemented high-assurance **Operational Velocity** visualization with stage latency heatmaps and settlement funnel analytics.
7. **Institutional Multi-Custody:** Implemented full-stack support for **BitGo Enterprise** and **Circle WaaS** via a dynamically pluggable **Stablecoin Factory** architecture.
8. **CIP-0056 Compliance:** Fully refactored the escrow lifecycle to use authoritative cryptographic locking and multi-actor co-signing per the CIP-0056 standard.

------------------------------------------------------------------------

## Phase 9 --- High-Assurance Identity & Deep Health (COMPLETE)

Goals: Establish production-grade OIDC bridge, automated infrastructure, and aggregated system diagnostics.

### Achievements

1. **Okta OIDC Bridge:** Implemented strict JWT validation against Okta JWKS. External identity assertions drive Just-In-Time (JIT) ledger provisioning.
2. **Identity-as-Code:** Automated the entire Okta stack (Apps, Servers, Users) using Terraform.
3. **Deep Health Aggregator:** Created a recursive diagnostic engine that aggregates the state of Postgres, Canton, and Oracle sub-systems with latency tracking.
4. **Directory Service:** Added a live counterparty discovery mechanism, allowing users to select authorized participants directly from the ledger.
5. **Production CLI:** Migrated the Go API to a professional Cobra/Viper structure, supporting environment-aware configuration via flags and YAML.

------------------------------------------------------------------------

## Phase 12 --- High-Assurance Secret Management (COMPLETE)

Goals: Replace local static credentials with authoritative cloud-native secret vending.

### Achievements

1. **Cloud-Native Vending:** Implemented dynamic secret fetching from **GCP Secret Manager**, eliminating the need for sensitive keys in local `.env` files.
2. **Terraform Integration:** Automated the provisioning of institutional secrets (`okta`, `bitgo`, `circle`) using Infrastructure-as-Code.
3. **Pre-Test Standard:** Established an authoritative "Local Pre-Test" pattern allowing developers to securely vend keys from the cloud during local development.

------------------------------------------------------------------------

## Phase 13 --- GCP Integration Testing (IN PROGRESS)

Goals: Implement specialized verification for real-world cloud environments (GCP/GKE).

### Tasks

- [ ] **Secret Manager Verification**: Implement `integration-gcp` tests for Secret Manager connectivity and vending logic.
- [ ] **Cloud SQL Schema Validation**: Build tests to verify Postgres schema and reference data deployment on Cloud SQL.
- [ ] **Networking Audit**: Implement connectivity checks for isolated participant namespaces (Bank/Buyer/Seller).
- [ ] **GKE Readiness**: Package production-grade Docker images for Artifact Registry and verify manifest parsing.

------------------------------------------------------------------------

## Phase 8 --- Distributed Sovereignty & Multi-Service Deployment (IN PROGRESS)

Goals: Transition to isolated service-per-node architecture.

### Achievements

1. **Structural Sovereignty:** Refactored Go API to support `PARTICIPANT_ID` locking, ensuring each container instance is authoritatively bound to a single ledger node.
2. **Tripartite Orchestration:** Implemented isolated Kubernetes manifests (`bank`, `buyer`, `seller`) for both Canton ledger nodes and Go API instances.
3. **High-Assurance Release Path:** Provisioned GCP Artifact Registry and established the `k8s/` guide for sovereign institutional deployment.

### Tasks

**Task 8.1: Service Containerization (COMPLETE)**

- [x] Refactor `main.go` to support single-node mode via environment variables.
- [x] Create `Dockerfile` for Go API and Frontend.
- [x] Implement health checks per service role.

**Task 8.2: Distributed API Routing (IN PROGRESS)**

- [ ] Implement API Gateway (Nginx/Envoy) with JWT-based routing logic.
- [ ] Configure cross-service telemetry for the distributed topology.

**Task 8.3: Compliance Validation**

- [ ] Perform "Node Isolation" security audit.
- [ ] Verify "Right to be Forgotten" patterns across distributed storage nodes.

------------------------------------------------------------------------

## Phase 7 --- Production Hardening & Compliance (SOC2 / Financial Grade)

Goals: Final security sweep to meet institutional regulatory and auditing standards.

### Tasks

**Task 7.1: Zero-Trust Networking (mTLS)**

- [ ] Provision **GCP Certificate Authority Service (CAS)** for internal cluster identity.
- [ ] Implement **Mutual TLS (mTLS)** enforcement across tripartite namespaces (`bank`, `buyer`, `seller`).
- [ ] Enforce TLS 1.3 for all external entry points via Ingress.

**Task 7.2: HSM & Key Management**

- [ ] Integrate **Google Cloud KMS (HSM-backed)** for ledger and API signing keys.
- [ ] Authoritatively migrate Okta and Stablecoin secrets to KMS-encrypted storage.

**Task 7.3: Immutable Auditing & Observability**

- [ ] Implement **Audit Log Sinks** for permanent, immutable transaction recording.
- [ ] Integrate **OpenTelemetry (OTEL)** for cross-service tracing and latency heatmaps.
- [ ] Export real-time metrics to Prometheus/Grafana for institutional monitoring.
