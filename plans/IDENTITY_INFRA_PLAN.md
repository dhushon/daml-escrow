# Integrated Identity & Health Strategy (Phase 9)

## 1. System Health Architecture (Deep Aggregation)

### 1.1 Single Source of Truth
The Go Backend API serves as the authoritative source for system health. The frontend MUST NOT query underlying infrastructure (Postgres, Ledger Nodes) directly.

### 1.2 Recursive Dependency Checks
The `/api/v1/health` endpoint performs a recursive check of its active dependencies:
- **Database (Postgres)**: Verified via `PingContext`.
- **Ledger (Canton)**: Verified via metadata package search.
- **Oracle**: Verified via configuration/secret validation.

### 1.3 Resilience & Timeouts
- **Timeout Requirement**: Every sub-service health check MUST implement a strict context timeout (default: 2 seconds).
- **Graceful Degradation**: If a non-critical dependency is slow or unreachable, the API MUST return a `DEGRADED` status instead of hanging or crashing.

---

## 2. Notification & Onboarding UX

### 2.1 Stateful Deep-Linking
To support notifications like "Your contract needs review," the platform implements a stateful redirect flow:
1.  **Capture**: The login gateway detects a `returnTo` parameter.
2.  **Persistence**: The destination is stored in `sessionStorage` across the OIDC redirect handshake.
3.  **Transit**: After successful Just-In-Time (JIT) provisioning, the user is automatically transited to the target record.

### 2.2 Secure Invitation Bridge (Private Onboarding)
To secure open email exchanges, the platform uses a cryptographic binding model:
- **Opaque Tokens**: Invitation links use high-entropy hashes. No PII (emails) or contract IDs are exposed in the URL.
- **Identity Binding**: The backend verifies that the `email` claim in the verified OIDC JWT matches the `inviteeEmail` stored on the ledger for that token.
- **Privacy Gating**: Contract amounts and sensitive terms are withheld from the UI until the identity-binding check is successfully completed.

---

## 3. Distributed Identity Management

### 3.1 Home Realm Discovery (HRD)
The system uses domain-based routing to select the correct Identity Provider:
- **Corporate Domains**: Routed via SAML to institutional providers.
- **Public Domains**: Routed via OIDC to the primary Okta instance.

### 3.2 Origin Tracking
Every identity assertion is tagged with its `origin_domain` (extracted from JWT or email suffix) to enable multi-tenant auditing and ledger-level partitioning.

---

## 4. Operational Baseline (Runtime)

### 4.1 Database Scaling
Due to the persistent connection pools required by multiple Canton participant nodes, mediators, and sequencers, the runtime database MUST be configured with a minimum `max_connections` of 500.

### 4.2 Security Guards
- **CORS**: The API must explicitly allow the frontend's origin and support the `X-Dev-User` header for development bypass testing.
- **OIDC Verification**: Every request must be verified against the provider's JWKS; insecure parsing is strictly prohibited in non-dev environments.

---

## 5. Next Steps (Roadmap)

### 5.1 Step 1: Security Hardening (Production Transition)
- **CORS Pruning**: Revert `AllowedOrigins: ["*"]` to a strict whitelist of the authorized frontend origin.
- **Strict Verification**: Remove `ParseUnverified` fallback. The API must 'fail closed' on any invalid OIDC signature.
- **Hardened Readiness**: Restore contract-specific template discovery in the ledger client to ensure commands only fire when the indexer is ready.

### 5.2 Step 2: Direct Identity Testing (Locked Configuration)
- **Lockdown Mode**: Configure testing environment for `ENVIRONMENT=dev` with `AUTH_BYPASS=false`.
- **Verified Handshake**: Validate the full OIDC handshake loop using real JWTs from a development Okta authorization server.
- **Provider Parity**: Ensure the identity bridge correctly handles both OIDC (Google/Okta) and SAML (Enterprise) assertions without bypass assistance.

### 5.3 Step 3: Advanced Discovery & 'Invite Now' Mechanics
- **Identity Directory**: Implement an API to enumerate or validate ledger-provisioned identities for counterparty selection.
- **Cryptographic Invites**: Implement the 'Invite Now' workflow where a user provides a target email, generating an opaque token cryptographically bound to that specific identity suffix.
- **Role Alignment**: Ensure that provisioned roles (Buyer, Seller, etc.) are strictly checked against ledger-level contract participation before allowing transitions.
