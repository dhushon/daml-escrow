# Daml & Canton Ledger Requirements (SDK 3.4.x)

## 1. Tooling & Package Management

- **MANDATORY:** Use **`dpm`** (Digital Asset Package Manager) for all contract operations. The legacy `daml` assistant is deprecated and MUST NOT be used for builds, tests, or deployments.
- **DPM Path:** Ensure `dpm` is in your path (typically `~/.dpm/bin/dpm`).
- **Multi-Package Builds:** Use `dpm build --all` from the `contracts/` directory to build the entire suite.

## 2. High-Assurance Escrow Lifecycle (Formal Model)

- **State Sequence:** ALL escrow contracts MUST follow the sequence: `DRAFT → FUNDED → ACTIVE → DISPUTED → PROPOSED → SETTLED`.
- **Tripartite Authority:**
  - **Issuer (Bank):** Signatory on ALL states. Must authorize final disbursement.
  - **Buyer & Seller:** Co-signers on terms (`DRAFT`) and settlement ratification (`PROPOSED`).
  - **Mediator:** Primary controller for condition confirmation and settlement proposals.
- **Audit Logging:** Every state transition MUST emit an `EscrowEvent` interface instance for auditability.

## 3. Authorization & Identity (Phase 5+)

- **Party ID Integrity:** Command submissions MUST use fully qualified Party IDs (e.g., `Buyer::1220abc...`).
- **User ID Scoping:** The `userId` MUST match the sanitized Daml User ID derived from the external IdP (e.g., `u-google-sub-123`).
- **JIT Provisioning Flow (SDK 3.4.x specific):**
    1. **Allocation:** `POST /v2/parties` with `{ "partyIdHint": "id", "displayName": "..." }`.
    2. **User Creation:** `POST /v2/users` with a nested user object.
    3. **Rights Granting:** `POST /v2/users/{id}/rights` requires `userId` and `identityProviderId` in the body.
- **Dynamic Resolution:** NEVER hardcode Party IDs. Always resolve via the `partyMap` cache refreshed from `/v2/parties`.

## 4. JSON API V2 Serialization

- **Nullary Constructors:** Data constructors with no fields MUST be represented as a plain string: `"payload": "ApproveMilestoneArg"`.
- **Zero-Argument Choices:** Choices with no parameters MUST use an empty object: `"choiceArgument": {}`.
- **Interface Exercises:** Targeting an interface choice MUST use the Interface Package ID and Template ID in the `ExerciseCommand`.

## 5. Stability & Performance

- **Dynamic Discovery:** Every ledger client MUST implement a `Discover(ctx)` phase to resolve Package IDs and Party IDs at runtime.
- **Hardcoding Prohibited:** Do NOT hardcode Package IDs or Party IDs. Use variables populated during discovery.
- **Timeout Management:** Complex transactions REQUIRE a minimum 120s client timeout.
- **Retry Strategy:** `GetEscrow` operations should use a minimum of 15 retries with 2s delays to account for Canton indexing latency.
