# Application Architecture

# Architecture Guardrails

1. Contract-First Design
2. Ledger as Source of Truth
3. Privacy by Default
4. Identity Bridge & JIT Provisioning (Phase 5+)
5. Stateless Services
6. Namespaced Configuration
7. Metadata & Oracle Guardrails
8. Self-Healing Lifecycle (Phase 5+)
9. Testing & Validation Hierarchy (Phase 6)
10. Distributed Identity & Routing
11. Tripartite Governance & Bilateral Consensus
12. **Off-Chain Draft Tunnel:** Utilizes a versioned Postgres/Cloud SQL layer (`draft_escrows`) for zero-latency bilateral negotiation before ledger commitment.
13. **Cloud-Native Persistence:** Migrated to GCP Cloud SQL with secret-vended DSNs for production-grade durability.
14. **Promotion Engine:** Authoritative transition from Ratified Draft -> DAML EscrowProposal/Invitation.
