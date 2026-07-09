# PLAN.md --- Stablecoin Escrow Platform

## Project Objective

Build a privacy-preserving **stablecoin escrow platform** using **DAML
smart contracts** and **Go-based services** for integration, APIs, and
orchestration.

---

## Phase 5 --- High-Assurance Lifecycle & Identity (COMPLETE)

Goals: Align contract dynamics with formal escrow standards and establish a production-grade identity bridge.

### Achievements

1. **Robust Escrow Lifecycle:** Refactored DAML and Go to support the standard `DRAFT → FUNDED → ACTIVE → DISPUTED → PROPOSED → SETTLED` sequence per `ESCROW-PROCESS.md`.
2. **Tripartite Authority:** Enforced co-signature requirements for Depositor, Beneficiary, and Issuer (Bank) on terms agreement and settlement ratification.
3. **Bilateral Dispute Resolution:** Implemented a two-stage mediated settlement flow requiring explicit ratification from both disputing parties.
4. **Dynamic Discovery:** Unified Go client to dynamically resolve `packageID` and `partyID` at runtime, ensuring portability across environments.
5. **JIT Identity Bridge:** Established automated mapping between Okta OIDC `sub` claims and unique DAML parties via the User Management API.
6. **Secure Onboarding:** Implemented `EscrowInvitation` flow with token-based "Claim" mechanics to bridge anonymous invitations to authenticated identities.

---

## Phase 6 --- Scaling & Distribution (COMPLETE)

Goals: Move to multi-participant topology and real stablecoin integration.

### Phase 6 Achievements

1. **Multi-Participant Infrastructure:** Created `contracts/DevNet/devnet.conf` and robust `devnet_init.canton` for distributed topology with deterministic propagation checks.
2. **Local Distributed Environment:** Added `docker-compose.distributed.yml` to enable local testing of separate Bank, Depositor, and Beneficiary nodes.
3. **Frontend Lifecycle Alignment:** Refactored `api.ts` and `EscrowCard.astro` to support the full high-assurance state machine (`DRAFT` → `FUNDED` → `ACTIVE` → `DISPUTED` → `PROPOSED` → `SETTLED`).
4. **Distributed Gateway:** Implemented `MultiLedgerClient` in Go for role-based node routing and resilient identity resolution across clusters.
5. **Testing Hierarchy:** Established three-tier testing strategy (Unit, Integration, Distributed) with reusable lifecycle logic and deterministic readiness checks.
6. **Advanced Analytics Dashboard:** Implemented high-assurance **Operational Velocity** visualization with stage latency heatmaps and settlement funnel analytics.
7. **Institutional Multi-Custody:** Implemented full-stack support for **BitGo Enterprise** and **Circle WaaS** via a dynamically pluggable **Stablecoin Factory** architecture.
8. **CIP-0056 Compliance:** Fully refactored the escrow lifecycle to use authoritative cryptographic locking and multi-actor co-signing per the CIP-0056 standard.

---

## Phase 9 --- High-Assurance Identity & Deep Health (COMPLETE)

Goals: Establish production-grade OIDC bridge, automated infrastructure, and aggregated system diagnostics.

### Phase 9 Achievements

1. **Okta OIDC Bridge:** Implemented strict JWT validation against Okta JWKS. External identity assertions drive Just-In-Time (JIT) ledger provisioning.
2. **Identity-as-Code:** Automated the entire Okta stack (Apps, Servers, Users) using Terraform.
3. **Identity Registry:** Created `docs/IDENTITY.md` authoritatively defining Contributor and Deployment Service principal models for SOC2 compliance.
4. **Deep Health Aggregator:** Created a recursive diagnostic engine that aggregates the state of Postgres, Canton, and Oracle sub-systems with latency tracking.
5. **Directory Service:** Added a live counterparty discovery mechanism, allowing users to select authorized participants directly from the ledger.
6. **Production CLI:** Migrated the Go API to a professional Cobra/Viper structure, supporting environment-aware configuration via flags and YAML.

---

## Phase 12 --- High-Assurance Secret Management (COMPLETE)

Goals: Replace local static credentials with authoritative cloud-native secret vending.

### Phase 12 Achievements

1. **Cloud-Native Vending:** Implemented dynamic secret fetching from **GCP Secret Manager**, eliminating the need for sensitive keys in local `.env` files.
2. **Terraform Integration:** Automated the provisioning of institutional secrets (`okta`, `bitgo`, `circle`) using Infrastructure-as-Code.
3. **Pre-Test Standard:** Established an authoritative "Local Pre-Test" pattern allowing developers to securely vend keys from the cloud during local development.

---

## Phase 13 --- GCP Integration Testing (COMPLETE)

Goals: Implement specialized verification for real-world cloud environments (GCP/GKE).

### Phase 13 Achievements

1. **High-Assurance GKE Foundation:** Provisioned cost-optimized zonal GKE clusters with Spot instances and institutional naming strategy (`-dev`).
2. **Unified API Gateway:** Implemented sovereign path-based routing (/bank, /depositor, /beneficiary) using Nginx Ingress to bridge tripartite namespaces.
3. **Specialized Cloud Verification:** Created the `integration_gcp` test suite to authoritatively verify Secret Manager connectivity and Cloud SQL schemas.

---

## Phase 8 --- Distributed Sovereignty & Multi-Service Deployment (COMPLETE)

Goals: Transition to isolated service-per-node architecture.

### Phase 8 Achievements

1. **Structural Sovereignty:** Refactored Go API to support `PARTICIPANT_ID` locking, ensuring each container instance is authoritatively bound to a single ledger node.
2. **Tripartite Orchestration:** Implemented isolated Kubernetes manifests (`bank`, `depositor`, `beneficiary`) for both Canton ledger nodes and Go API instances.
3. **High-Assurance Release Path:** Provisioned GCP Artifact Registry and established the `k8s/` guide for sovereign institutional deployment.

---

## Phase 10 --- Institutional Pilot & Verification (IN PROGRESS)

Goals: Authoritative release and live verification under the vdatacloudai.com domain.

### Phase 10 Achievements

1. **Authoritative Release:** Successfully pushed production-grade sovereign images to the GCP Artifact Registry.
2. **Live Orchestration:** Deployed the full tripartite stack (Bank, Depositor, Beneficiary) to the live GKE cluster.
3. **Sovereign Perimeter:** Established the `api.vdatacloudai.com` entrypoint via Global Static IP and Let's Encrypt TLS 1.3.

### Tasks

- [ ] **Internal mTLS Enforcement (Production Only):** Finalize sidecar certificate injection via GCP CAS.
- [ ] **Live Health Audit:** Perform deep health verification across all tripartite namespaces.
- [ ] **Domain Verification:** Confirm authoritative mapping and SSL termination at api.vdatacloudai.com.

---

## Phase 11 --- Sovereign Negotiation & Off-Chain Drafts (COMPLETE)

Goals: Implement cost-efficient institutional negotiation via intermediate draft tables and secure counterparty invitations.

### Achievements

1. **Off-Chain Draft Intermediate:** Utilized the Postgres `user_config` database to authoritatively house versioned negotiations, enabling low-cost bilateral iteration before ledger commitment.
2. **Invitation & Association Bridge:** Implemented registration-code-based onboarding to bridge unprovisioned email identities to real ledger principals.
3. **Bilateral Consensus Logic:** Built a negotiation cockpit where Depositors and Beneficiaries iterate on terms, authoritatively resetting approvals upon amendment.
4. **Authoritative Ledger Promotion:** Built the promotion engine that authoritatively translates ratified drafts into active DAML contracts with immediate stablecoin locking.

---

## Phase 15 --- Intelligent Ingest & Legal AI Alignment (COMPLETE)

Goals: Authoritatively bridge legacy legal prose with DAML smart contracts using AI-native ingestion.

### Achievements

1. **AI Ingest Engine:** Integrated **Gemini-2.0-flash** to authoritatively classify typology and extract structured economic terms from multi-page agreements (PDF/PNG/TIFF).
2. **Multi-Part File Sequencing:** Implemented a high-fidelity frontend **File Sequencer** allowing users to drag-and-drop file identifiers into their natural page order before processing.
3. **Contract Typology:** Introduced industry-specific JSON schemas (**Import/Export**, **Real Estate**, **Grants**, **Corporate**) to authoritatively validate extracted terms before ledger commitment.
4. **Enriched Identity Modeling:** Expanded the institutional directory to capture **Titles**, **Corporate Affiliations**, and **KYC Status**, authoritatively linking signatories to real system identities.

---

## Phase 16 --- High-Assurance Read-Through Storage & Mirroring (COMPLETE)

Goals: Implement production-grade document privacy and sovereign storage mirroring.

### Achievements

1. **Valet Storage Pattern:** Implemented a high-assurance object storage model using **GCS (Production)** and **MinIO (Local)** with zero public access.
2. **Read-Through Lazy Mirroring:** Built the authoritative distribution engine where document blobs are automatically "cached" in a party's local vault upon retrieval.
3. **Dynamic Re-signing:** Implemented backend-authorized **Presigned URLs** that are dynamically re-signed for a user's specific vault for every fetch, authoritatively enforcing privacy.
4. **SSE-KMS Encryption:** Enforced hardware-backed **Encryption at Rest** (SSE-KMS / GCP CMEK) for all institutional document blobs.
5. **Authoritative Object Tagging:** Implemented S3/GCS tagging to link blobs to `contract-id`, `depositor`, and `beneficiary` identities, enabling cross-vault searchability.

---

## Phase 7 --- Production Hardening & Compliance (SOC2 / Financial Grade) (IN PROGRESS)

Goals: Final security sweep to meet institutional regulatory and auditing standards.

### Phase 7 Achievements

1. **Hardware-Backed Security:** Provisioned GCP KMS infrastructure and implemented asymmetric oracle verification, establishing the definitive HSM-protected root of trust.
2. **High-Assurance Observability:** Implemented full-stack OpenTelemetry (OTEL) integration with Jaeger (tracing), Prometheus (metrics), and pre-provisioned Grafana dashboards.
3. **Sovereign Telemetry:** Enabled granular performance tracking at the System, Account, and Contract level, tagged by authenticated institutional identities.

### Task 7.1: Zero-Trust Networking (mTLS)

- [x] **Local Development PKI:** Provision local self-signed cert-manager CA issuer for Minikube.
- [ ] **GCP CAS (Production Only):** Provision GCP Certificate Authority Service (CAS) for internal GKE cluster identity.
- [ ] **mTLS Enforcement (Production Only):** Implement Mutual TLS (mTLS) enforcement across tripartite namespaces (`bank`, `depositor`, `beneficiary`).
- [ ] **TLS 1.3 Perimeter:** Enforce TLS 1.3 for all external entry points via Ingress.

### Task 7.2: HSM & Key Management (COMPLETE)

- [x] Integrate **Google Cloud KMS (HSM-backed)** for ledger and API signing keys.
- [x] Authoritatively migrate Oracle and Stablecoin triggers to asymmetric signing.

### Task 7.3: Immutable Auditing & Observability (COMPLETE)

- [x] Implement **Audit Log Sinks** for permanent, immutable transaction recording.
- [x] Integrate **OpenTelemetry (OTEL)** for cross-service tracing and latency heatmaps.
- [x] Export real-time metrics to Prometheus/Grafana for institutional monitoring.

---

## Phase 17 --- Complex Escrow Topologies (PLANNED)

Goals: Extend the DAML contract model beyond bilateral escrow to support multi-party consent, milestone-independent settlement, and nested contract composition, per `ESCROW-PROCESS.md` DIRECTIVES 10-12.

### Phase 17 Tasks

- [ ] **Multi-Party Consent Model:** Generalize `Buyer`/`Seller` template parameters into `BuyerSet`/`SellerSet` with a configurable `ConsentThreshold`, preserving today's unanimous single-party behavior as the default case.
- [ ] **Weighted Disbursement Split:** Implement `WeightedSplit` on `SettlementTerms` so multi-beneficiary payouts do not default to an equal split unless explicitly configured that way.
- [ ] **Milestone-Independent Release:** Replace the contract-wide `ConfirmConditions` gate with per-`MilestoneId` verification and release, so unrelated tranches can clear independently.
- [ ] **Nested Escrow Composition:** Implement `SpawnChildEscrow` and `SettleParent`, with parent-child party authorization checks to prevent privilege escalation through nesting.

---

## Phase 18 --- Multi-Rail Settlement Orchestration (PLANNED)

Goals: Add a fiat disbursement leg alongside the existing stablecoin path, without changing contract-level authority logic, per `ESCROW-PROCESS.md` DIRECTIVE 14 and `FIAT-SETTLEMENT.md`.

### Phase 18 Tasks

- [ ] **RailRouter Package:** Build `internal/railrouter` as the dispatch layer behind `Disburse`, routing to either the existing `StablecoinFactory` or a new `FiatProvider` implementation.
- [ ] **FiatProvider Interface:** Define the minimal Go interface (`InitiateTransfer`, `GetStatus`, `RegisterWebhook`) so the fiat backend stays swappable rather than hardcoded to one vendor.
- [ ] **Payments Orchestration Integration:** Implement the first `FiatProvider` against a payments orchestration API supporting ACH, RTP, FedNow, and wire, evaluated against the existing pluggable-custody pattern already used for BitGo/Circle.
- [ ] **FIAT_PENDING Sub-State:** Add the intermediate state to the DAML state diagram and frontend `EscrowCard.astro`, distinguishing ledger-recorded intent from confirmed off-ledger settlement.
- [ ] **Webhook Confirmation Endpoint:** Extend the existing Oracle webhook pattern to accept fiat settlement confirmations, closing the loop via `ConfirmFiatSettlement`.
- [ ] **Dashboard Extension:** Add a fiat-settlement latency band to the Operational Velocity stage-duration heatmap, kept separate from on-chain confirmation timing.

---

## Phase 19 --- Yield Accrual & Dispute Escalation (PLANNED)

Goals: Close two gaps identified against real-world escrow complexity, per `ESCROW-PROCESS.md` DIRECTIVES 13 and 15.

### Phase 19 Tasks

- [ ] **Accrual Policy Field:** Add `AccrualPolicy`, `AccrualRateBps`, and `AccrualBeneficiary` to the contract terms block, immutable after FUNDED.
- [ ] **Invariant Revision:** Update INVARIANT I1 to distinguish escrowed principal from accrued yield, and to account for an explicit `TopUp` choice if CDP-backed escrow moves forward.
- [ ] **Arbitration Tier:** Implement `EscalateToArbitration` and `RenderArbitrationDecision` as a binding exit path for disputes that exceed the mediation rejection threshold.
- [ ] **Regulatory Doc Update:** Extend `REGULATORY_CONFORMANCE.md` to cover payout-side AML/KYC posture now that fiat rails introduce external-account exposure not present in the stablecoin-only model.

---

## Phase 20 --- Frontend Rework for Complex Escrow (PLANNED)

Goals: Bring the Astro frontend up to parity with the reworked contract model, per `FRONTEND-PROCESS.md`.

### Phase 20 Tasks

- [ ] **State Machine Component Update:** Extend `EscrowCard.astro` and `api.ts` to render ARBITRATION and FIAT_PENDING states, and to stop assuming a single Buyer/Seller party.
- [ ] **Multi-Party Consent View:** Build a party-set roster component showing each BuyerSet/SellerSet member's ratification status against ConsentThreshold.
- [ ] **Milestone Board:** Replace the single condition-status indicator with a per-milestone board supporting independent verify/release actions and per-milestone dispute flags.
- [ ] **Parent/Child Navigation:** Add a rollup view for parent escrows showing child contract status and the active AggregationRule.
- [ ] **Arbitration Flow UI:** Surface the escalation path after repeated rejections, and a distinct binding-decision view for Arbitrator actions.
- [ ] **Rail Selection & FIAT_PENDING UX:** Add a settlement rail selector at settlement-terms entry, and a pending-confirmation state for fiat disbursements distinct from on-chain settlement.
- [ ] **Dashboard Extension:** Extend the Operational Velocity dashboard with milestone-level funnel stages and a separate fiat-settlement latency band.

---

*End of PLAN.md*