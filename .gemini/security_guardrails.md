# Security Guardrails

Security rules for high-assurance escrow and financial contract systems.

## 1. Critical Security Principles
1.  **Funds safety first:** Every state transition involving funds must be deterministic and ledger-verified.
2.  **Least privilege:** Users and services are granted the absolute minimum permissions required.
3.  **Explicit authorization:** No operation is permitted without verified identity and valid scope.
4.  **Auditability:** Every significant action must be recorded on the immutable ledger or authorized audit log.
5.  **Fail-safe defaults:** Authorization checks must default to "Deny All".

## 2. Identity & Authorization (Phase 5+)
- **Authoritative IdP:** Google Cloud Identity Platform (GCIP) is the primary source of identity.
- **Token Verification:** Every API request MUST include a valid OIDC JWT. Use standard libraries (`golang-jwt`) for verification.
- **Scoped Permissions:** API endpoints MUST enforce granular scopes:
    - `escrow:read`: View contracts, proposals, and invites.
    - `escrow:write`: Propose agreements or create invites.
    - `escrow:accept`: Execute legal acceptance or fund release.
    - `system:admin`: Perform settlements, resolution, or system config.
- **JIT Integrity:** Automated provisioning MUST follow strict sanitization rules for User IDs.

## 3. Smart Contract Rules
Contracts must enforce:
-   **Explicit controllers:** No "backdoor" or "admin" overrides on financial choices.
-   **No implicit approvals:** All counterparty agreements must be signed explicitly.
-   **Deterministic settlement:** Funds can only be released to verified recipients via the Settlement interface.

## 4. Service Security
Services must:
-   **Validate all input:** Use strict typing and schema validation.
-   **Verify identities:** Never trust client-provided IDs; always derive from verified JWT claims.
-   **Avoid trust in client state:** Always re-fetch contract state from the ledger before executing an action.

## 5. Secrets Management
Secrets (API keys, Webhook secrets, DSNs) must never appear in code, config files, or logs. Use environment variables or Cloud Secret Manager.

## 6. Oracle Security
-   **HMAC Verification:** Webhooks MUST be authenticated using HMAC-SHA256 signatures.
-   **State Check:** Trigger logic MUST fetch current ledger state to verify milestone indices before acting.
