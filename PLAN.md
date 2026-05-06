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

## Phase 13 --- GCP Integration Testing (COMPLETE)

Goals: Implement specialized verification for real-world cloud environments (GCP/GKE).

### Achievements

1. **High-Assurance GKE Foundation:** Provisioned cost-optimized zonal GKE clusters with Spot instances and institutional naming strategy (`-dev`).
2. **Unified API Gateway:** Implemented sovereign path-based routing (/bank, /buyer, /seller) using Nginx Ingress to bridge tripartite namespaces.
3. **Specialized Cloud Verification:** Created the `integration_gcp` test suite to authoritatively verify Secret Manager connectivity and Cloud SQL schemas.

------------------------------------------------------------------------

## Phase 8 --- Distributed Sovereignty & Multi-Service Deployment (COMPLETE)

Goals: Transition to isolated service-per-node architecture.

### Achievements

1. **Structural Sovereignty:** Refactored Go API to support `PARTICIPANT_ID` locking, ensuring each container instance is authoritatively bound to a single ledger node.
2. **Tripartite Orchestration:** Implemented isolated Kubernetes manifests (`bank`, `buyer`, `seller`) for both Canton ledger nodes and Go API instances.
3. **High-Assurance Release Path:** Provisioned GCP Artifact Registry and established the `k8s/` guide for sovereign institutional deployment.

------------------------------------------------------------------------

## Phase 7 --- Production Hardening & Compliance (SOC2 / Financial Grade) (IN PROGRESS)

Goals: Final security sweep to meet institutional regulatory and auditing standards.

### Achievements

1. **Hardware-Backed Security:** Provisioned GCP KMS infrastructure and implemented asymmetric oracle verification, establishing the definitive HSM-protected root of trust.
2. **High-Assurance Observability:** Implemented full-stack OpenTelemetry (OTEL) integration with Jaeger (tracing), Prometheus (metrics), and pre-provisioned Grafana dashboards.
3. **Sovereign Telemetry:** Enabled granular performance tracking at the System, Account, and Contract level, tagged by authenticated institutional identities.

### Tasks

**Task 7.1: Zero-Trust Networking (mTLS)**

- [ ] Provision **GCP Certificate Authority Service (CAS)** for internal cluster identity.
- [ ] Implement **Mutual TLS (mTLS)** enforcement across tripartite namespaces (`bank`, `buyer`, `seller`).
- [ ] Enforce TLS 1.3 for all external entry points via Ingress.

**Task 7.2: HSM & Key Management (COMPLETE)**

- [x] Integrate **Google Cloud KMS (HSM-backed)** for ledger and API signing keys.
- [x] Authoritatively migrate Oracle and Stablecoin triggers to asymmetric signing.

**Task 7.3: Immutable Auditing & Observability (COMPLETE)**

- [x] Implement **Audit Log Sinks** for permanent, immutable transaction recording.
- [x] Integrate **OpenTelemetry (OTEL)** for cross-service tracing and latency heatmaps.
- [x] Export real-time metrics to Prometheus/Grafana for institutional monitoring.
