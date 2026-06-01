# Phase 17: CIP-0103 Wallet Integration & Progressive Custody

## Objective
Integrate the Canton Network standard for wallet and app interoperability (**CIP-0103**) using the `@canton-network/dapp-sdk`. Establish a **Progressive Custody** architecture that supports both traditional custodial users (via Okta) and self-sovereign users (via connected wallets) within a unified API security model.

## Background & Motivation
Currently, the platform acts as a custodial fiduciary for all participants, utilizing Okta for authentication and signing transactions on behalf of users via the backend's system identities. To support institutional participants who require self-sovereign key management (using external wallets like Splice Wallet), we must adopt the CIP-0103 standard. This standard decouples the dApp from specific wallet implementations.

To avoid alienating non-crypto-native users, we will implement a "Dual Path" (Progressive Custody) model.

## Proposed Solution

### 1. Progressive Custody Architecture (Dual Path)
The platform will support two modes of operation simultaneously:
*   **Path A (Custodial)**: Users log in via Okta. The Go backend maintains their mapping and automatically signs/submits DAML commands to the ledger.
*   **Path B (Self-Sovereign/CIP-0103)**: Users connect an external Canton-compatible wallet. The backend generates unsigned command payloads, and the frontend SDK requests the user's wallet to sign and submit them.

### 2. Wallet-as-Identity Authentication
To grant self-sovereign users access to off-chain data (Drafts, Ingest PDFs), we implement a Challenge/Response auth flow:
1.  **Challenge**: Frontend requests a cryptographically secure, single-use nonce from the backend.
2.  **Sign**: Frontend uses the dApp SDK to ask the wallet to sign the nonce.
3.  **Verify**: Backend verifies the signature against the wallet's Public Key/Party ID.
4.  **Session**: Backend issues a standard JWT session token containing the `auth_method: "wallet"` and the verified `party_id`.

### 3. Strict Session-to-Wallet Binding
To prevent "context bleed" (where a user acts on Company A's drafts using Company B's wallet):
*   A session is irrevocably bound to the specific Wallet/Party ID verified at login.
*   The frontend will monitor the dApp SDK for `accountChanged` or `disconnect` events.
*   If the wallet context changes, the frontend will immediately invalidate the local session and force a hard redirect to the `/login` page. Hot-swapping wallets mid-session is prohibited.

### 4. Payload Delegation
The Go backend remains the authoritative source of business logic. When a Self-Sovereign user initiates a ledger action (e.g., "Accept Proposal"):
*   The backend validates the request.
*   Instead of executing the command, the backend serializes the DAML command payload and returns it to the frontend.
*   The frontend passes this payload to `sdk.prepareExecute()` and requests the wallet to sign/submit.

## Implementation Plan

### Phase 17.1: Backend Auth Extensions (Go)
1.  Add endpoints for Wallet Auth: `GET /api/v1/auth/nonce` and `POST /api/v1/auth/wallet/verify`.
2.  Implement cryptographic signature verification in the `IdentityService` to validate wallet signatures.
3.  Update the JWT generator to include `auth_method` (custodial vs wallet).

### Phase 17.2: Payload Generation Services (Go)
1.  Refactor the `EscrowService` methods (Propose, Fund, Activate, etc.) to support a "Dry Run" or "Generate Payload" mode.
2.  When a request originates from a `wallet` authenticated session, return the serialized DAML commands instead of submitting them to the Canton API.

### Phase 17.3: Frontend dApp Integration (Astro)
1.  Install `@canton-network/dapp-sdk` in the `/frontend` project.
2.  Update `/login` with a "Connect Wallet" option alongside the Okta flow.
3.  Implement the Wallet Challenge/Response flow in `auth.ts`.
4.  Implement event listeners to destroy the session on wallet change.
5.  Update UI action buttons to conditionally handle the "Payload Delegation" flow (receive payload -> trigger wallet signature).

## Verification & Testing
*   **Wallet Auth Test**: Verify that a simulated wallet signature successfully yields a valid platform JWT.
*   **Context Isolation Test**: Connect Wallet A, establish session, switch to Wallet B in the SDK, verify immediate session termination and redirect.
*   **Payload Delegation Test**: Verify that a "Propose Escrow" action by a wallet user returns a valid DAML command payload without altering ledger state directly from the backend.

## Migration & Rollback
*   The `AuthMiddleware` remains fully backwards compatible with existing Okta-issued tokens.
*   Custodial users will experience no change in workflow.
