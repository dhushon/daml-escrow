# High-Assurance Off-Chain Negotiation Plan (Phase 11)

This document authoritatively defines the functional flow for off-chain negotiation and secure ledger promotion.

## 1. The "Draft Tunnel" Architecture

To minimize ledger transaction costs and institutional friction, negotiation occurs in an intermediate persistence layer (Postgres) before authoritative commitment to the Canton Ledger.

### Negotiation Lifecycle
1.  **DRAFT**: Initiator creates a draft agreement. Counterparty can be a registered ID or a raw email address.
2.  **NEGOTIATION**: Parties propose changes to milestones, amounts, or mediators. Each change authoritatively resets all existing approvals.
3.  **RATIFIED**: All three parties (Buyer, Seller, Mediator) have clicked "Approve" on the identical draft version.
4.  **PROMOTED**: A single high-assurance trigger commits the draft to the ledger and locks stablecoin holdings.

## 2. Invitation & Association Bridge

This flow allows institutional commerce to proceed before all parties have finalized their chain identity.

| Stage | Action | Authoritative Result |
| :--- | :--- | :--- |
| **Invite** | Initiator sends draft to `invitee@target.com`. | Draft is created in DB with `counterparty_email` placeholder. |
| **Onboard** | Invitee clicks link and logs in via Okta. | JIT provisioning allocates a `damlPartyId`. |
| **Claim** | System matches email to `draft.counterparty_email`. | The `counterparty_id` is updated to the new chain identity. |

## 3. Contract-First Data Modeling

We define our institutional structures in a centralized schema repository.

*   **Location**: `architecture/schemas/*.json` (e.g., `escrow.json`, `grants.json`).
*   **Role**: These schemas inform:
    *   **UX**: Dynamic form generation and validation.
    *   **API**: Strict DTO validation and error reporting.
    *   **Ledger**: The full record is stored as a JSON blob in the `metadata` field, ensuring 100% data fidelity.

## 4. Promotion & Finality

Promotion to the ledger is the definitive high-assurance gate.

1.  **Identity Verification**: The API confirms all 3 parties have valid `damlPartyId`s.
2.  **Consensus Check**: Verifies that the draft hash matches the version approved by all parties.
3.  **Chain Filing**: Executes a single Daml transaction to create the `ACTIVE` escrow and authoritatively lock funds.
