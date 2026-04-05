# Daml & Canton Ledger Requirements (SDK 3.4.x)

## 1. Tooling & Package Management

- **MANDATORY:** Use **`dpm`** (Digital Asset Package Manager) for all contract operations. The legacy `daml` assistant is deprecated and MUST NOT be used for builds, tests, or deployments.
- **DPM Path:** Ensure `dpm` is in your path (typically `~/.dpm/bin/dpm`).
- **Multi-Package Builds:** Use `dpm build --all` from the `contracts/` directory to build the entire suite.

## 2. Distributed Multi-Participant Topology (Phase 6)

- **Node Specialization:** In a distributed environment, parties are hosted on specific nodes (e.g., `CentralBank` on `bank` node, `Buyer` on `buyer` node).
- **Synchronizer Sovereignty:** All cross-node interactions MUST be authorized on the synchronizer topology store.
- **Tripartite Authorization:** Topology mappings (`party_to_participant_mappings`) MUST be explicitly proposed and fully authorized by all involved participants to enable command submission.
- **Deterministic Propagation:** NEVER rely on fixed `sleep` calls for topology. Use algorithmic verification:
    - **Canton Script:** Loop `topology.party_to_participant_mappings.list` until all expected parties are visible on all nodes.
    - **Go Client (Readiness):** The ledger API is eventually consistent. Upon startup, the client MUST perform a **Deterministic Readiness Check**:
        - Query the `EscrowProposal` template (or similar core template) via `/v2/query`.
        - Implement a 10-retry loop with 5s sleep between attempts.
        - Fail if the template is not indexed after the full cycle.
    - **Go Client (Discovery):** Implement `Discover()` with exponential backoff, checking for both Package IDs and all required Core Party IDs.

## 3. Testing Hierarchy & Determinism

- **Unit Tests (`go test ./...`):** Logic tests using mocks. MUST NOT require Docker or Network access.
- **Local Integration (`-tags=integration`):** Single-node Sandbox testing. Validates contract logic against a real ledger API.
- **Distributed Integration (`-tags=distributed`):** Multi-node Canton topology testing. Validates routing, cross-node authorization, and topology propagation.
- **Infrastructure Isolation:** Use Go build tags to ensure `go test ./...` remains fast and infra-agnostic.

## 4. Multi-Node Routing & Identity Bridge

- **Identity Bridge:** The platform uses an OIDC-to-Ledger mapping strategy. External JWT assertions are cryptographically verified and used to drive Just-In-Time (JIT) ledger provisioning.
- **Identity Sanitization:** External subjects MUST be sanitized for ledger compatibility:
    - Replace `|`, `@`, `.`, and `_` with `-`.
    - Apply lowercase transformation.
    - Prefix with `u-` (e.g., `hushon@gmail.com` -> `u-hushon-gmail-com`).
- **Scope-to-Permission Mapping:** OIDC scopes (e.g., `system:admin`) are mapped directly to ledger-level rights (e.g., `actAs CentralBank`). This ensures that external identities carry cryptographically enforceable authority on the synchronizer.
- **Home Realm Discovery (HRD):** Domain-based routing determines the correct IdP (Okta/SAML) and tracks identity origin via the `origin_domain` claim.
- **Routing Logic:** The `MultiLedgerClient` MUST route commands to the node hosting the primary submitter. 
- **Identity Redundancy:** `GetIdentity` MUST probe all nodes in the cluster to resolve users, as users may be provisioned on different participants depending on their role/email domain.
- **Detailed Documentation:** For complete details on the identity bridge, JIT provisioning, and HRD strategies, see [**IDENTITY.md**](../IDENTITY.md).

## 5. High-Assurance Escrow Lifecycle

- **State Sequence:** ALL escrow contracts MUST follow: `DRAFT → FUNDED → ACTIVE → DISPUTED → PROPOSED → SETTLED`.
- **Signatory Rules:**
  - **Issuer (Bank):** Signatory on ALL states. 
  - **Seller Acceptance:** Added `SellerAcceptedProposal` state to handle multi-stage co-signing without requiring all parties online simultaneously.
- **Audit Logging:** Every state transition MUST emit an `EscrowEvent` interface instance.

## 6. JSON API V2 Serialization

- **Nullary Constructors:** Data constructors with no fields MUST be represented as a plain string: `"payload": "ApproveMilestoneArg"`.
- **Zero-Argument Choices:** Choices with no parameters MUST use an empty object: `"choiceArgument": {}`.
- **ActiveAtOffset:** ACS queries (`/v2/query`) MUST include `"activeAtOffset"` derived from the previous transaction response to ensure consistency.

## 8. Critical Workarounds & Nuances (Phase 6 Findings)

### A. Canton 3.x Console API
- **User Creation:** When creating users via the console (e.g., in `init.canton`), always use named parameters. `users.create(id = "uid", primaryParty = Some(party))` is required to distinguish from the `actAs` Set.
- **Topology Propose:** `party_to_participant_mappings.propose` must be called with `ParticipantPermission.Submission` and `mustFullyAuthorize = true` to enable cross-node command submission.

### B. JSON API V2 Consistency
- **Offset Management:** The V2 API is eventually consistent. To guarantee "read-your-own-writes," capture the `offset` field from every command response. In subsequent ACS queries, pass this as `activeAtOffset` (as a string) to ensure the indexer has caught up.
- **Package Discovery:** Package name resolution is disabled in V2 for performance. Always prefer `ledger-state.json` (generated by `dpm` or `make sync`) over active discovery.

### C. Multi-Participant Identity
- **Distributed Identity:** In a 3-node cluster, `GetIdentity` must probe ALL nodes. A user might be provisioned on the Buyer node but the query hits the Bank node.
- **Connection Resiliency:** The `MultiLedgerClient` should implement a "Node Routing" strategy: route to the participant hosting the `primaryParty` of the user.

