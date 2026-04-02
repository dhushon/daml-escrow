# Identity Strategy & Decentralized Onboarding

## Overview
This document defines the transition from a "Sandbox" identity model to a production-grade model driven by **Okta OIDC**, **Daml User Management**, and **Scoped Authorization**. The platform implements a high-assurance "Identity Bridge" that maps external identity assertions directly to Canton ledger permissions.

---

## 1. The Identity Mapping (OIDC to Daml)

We establish an "Identity Bridge" in the Go backend that translates external authentication into ledger authorization.

### Data Flow:
1.  **Authentication:** User logs into the Astro frontend. The system performs **Home Realm Discovery (HRD)** based on the email domain to select the correct provider (e.g., Okta for `gmail.com`, SAML for `bank.com`).
2.  **Credential:** The Identity Provider issues a JWT containing a `sub` (unique subject ID), `email`, and `scp` (scopes).
3.  **Validation:** The Go backend cryptographically verifies the JWT signature against the provider's **JWKS endpoint** using the `go-oidc` library.
4.  **Authorization:** The verified `sub` is sanitized and mapped to a **Daml User ID**.
5.  **JIT Provisioning:** If the user is new, the backend automatically allocates a new **Daml Party** and provisions a User with appropriate `actAs` rights derived from the JWT scopes.

---

## 2. JIT Provisioning & Permission Mapping

During implementation, the following high-assurance requirements for the Daml JSON API (V2) were integrated into the Just-In-Time (JIT) flow:

### A. Party Allocation (`POST /v2/parties`)
- **Field:** Use `partyIdHint` (not `identifierHint`).
- **Response:** The allocated party is found at `partyDetails.party`.

### B. User Creation (`POST /v2/users`)
External subjects are sanitized (e.g., `hushon@gmail.com` -> `u-hushon-gmail-com`) to meet Daml identifier requirements.
```json
{
  "user": {
    "id": "u-sanitized-sub",
    "primaryParty": "allocated-party-id",
    "isDeactivated": false,
    "identityProviderId": ""
  }
}
```

### C. Rights Mapping (Directive 11)
The backend translates OIDC scopes into cryptographic ledger rights:
- **Scope `system:admin`:** Automatically grants `actAs` rights for the **CentralBank** party, enabling institutional oversight and disbursement actions.
- **Default:** Every provisioned user is granted `actAs` rights for their own primary party.

---

## 3. Home Realm Discovery (HRD) Strategy

To support multi-tenancy and enterprise federation, the platform uses a domain-based discovery mechanism:

1.  **Lookup:** The frontend calls `/api/v1/auth/discover?email=user@domain.com`.
2.  **Mapping:** The `IdentityService` consults `config/identity_providers.yaml` to return the correct OIDC `issuer` or SAML `loginUrl`.
3.  **Origin Tracking:** The backend extracts the `origin_domain` claim from the JWT (with a fallback to the email suffix) to track which company asserted the user's identity.

---

## 4. Authorization Matrix (Scopes)

| Role | Scopes | Ledger Rights | Allowed Operations |
| :--- | :--- | :--- | :--- |
| **Viewer** | `escrow:read` | `readAs` | List/Get Escrows, Proposals, Invitations. |
| **Contributor** | `escrow:write` | `actAs (Self)` | Propose Escrow, Create Invitation. |
| **Participant** | `escrow:accept` | `actAs (Self)` | Accept Proposal, Ratify Settlement. |
| **Admin** | `system:admin` | `actAs (Bank)` | Settle (Bank), Activate, Disburse. |

---

## 5. Dynamic Identity & Package Discovery

To maintain high assurance across contract upgrades and environment resets, the platform avoids hardcoding cryptographic identifiers. Instead, it employs a **Discovery Phase** at startup.

### Strategy:
1.  **Logical Mapping:** The backend maintains knowledge of "Logical Names" (e.g., package `stablecoin-escrow`, party `Buyer`).
2.  **Runtime Resolution:** Upon connection, the `ledgerClient.Discover(ctx)` method is executed.
3.  **Package Sync:** The system queries the ledger's package registry to resolve the current content-hashes (Package IDs) for interfaces and implementations.
4.  **Party Sync:** Cryptographic Party IDs are resolved via the User Management API and identifier hints, populating a high-speed local cache (`partyMap`).

This ensures that the Go backend always interacts with the exact versions of the contracts currently active on the Synchronizer.
