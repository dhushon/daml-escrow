# Identity Strategy & Decentralized Onboarding

## Overview
This document defines the transition from a "Sandbox" identity model to a production-grade model driven by **Google Cloud Identity (Identity Platform)**, **Daml User Management**, and **Scoped Authorization**.

---

## 1. The Identity Mapping (Google to Daml)

We establish an "Identity Bridge" in the Go backend that translates external authentication into ledger authorization.

### Data Flow:
1.  **Authentication:** User logs into the Astro frontend via Google OIDC.
2.  **Credential:** Google issues a JWT containing a `sub` (unique subject ID) and `email`.
3.  **Authorization:** The Go backend verifies the JWT and resolve the `sub` to a **Daml Party ID**.
4.  **Scopes:** Internal permissions are derived from the JWT's claims or mapped via internal policy.

---

## 2. JIT Provisioning (SDK 3.4.x Nuances)

During implementation, the following high-assurance requirements for the Daml JSON API (V2) were identified:

### A. Party Allocation (`POST /v2/parties`)
- **Field:** Use `partyIdHint` (not `identifierHint`).
- **Response:** The allocated party is found at `partyDetails.party`.

### B. User Creation (`POST /v2/users`)
The payload must be a nested `user` object with specific compliance fields:
```json
{
  "user": {
    "id": "u-sanitized-sub",
    "primaryParty": "party-id",
    "isDeactivated": false,
    "identityProviderId": ""
  }
}
```

### C. Rights Granting (`POST /v2/users/{id}/rights`)
The request body MUST include `userId` and `identityProviderId` even though the ID is in the URL:
```json
{
  "userId": "u-id",
  "identityProviderId": "",
  "grant": [{ "type": "actAs", "party": "party-id" }]
}
```

---

## 3. Authorization Matrix (Scopes)

| Role | Scopes | Allowed Operations |
| :--- | :--- | :--- |
| **Viewer** | `escrow:read` | List/Get Escrows, Proposals, Invitations. |
| **Contributor** | `escrow:write` | Propose Escrow, Create Invitation, Webhook actions. |
| **Participant** | `escrow:accept` | Accept Proposal, Claim Invitation, Release Funds. |
| **Admin** | `system:admin` | Settle (Bank), Resolve Dispute (Mediator), Metrics, Config. |

---

## 5. Dynamic Identity & Package Discovery

To maintain high assurance across contract upgrades and environment resets, the platform avoids hardcoding cryptographic identifiers. Instead, it employs a **Discovery Phase** at startup.

### Strategy:
1.  **Logical Mapping:** The backend maintains knowledge of "Logical Names" (e.g., package `stablecoin-escrow`, party `Buyer`).
2.  **Runtime Resolution:** Upon connection, the `ledgerClient.Discover(ctx)` method is executed.
3.  **Package Sync:** The system queries the ledger's package registry to resolve the current content-hashes (Package IDs) for interfaces and implementations.
4.  **Party Sync:** Cryptographic Party IDs are resolved via the User Management API and identifier hints, populating a high-speed local cache (`partyMap`).

This ensures that the Go backend always interacts with the exact versions of the contracts currently active on the Synchronizer.
