# Identity Strategy & Decentralized Onboarding

## Overview
This document defines the transition from a "Sandbox" identity model (where parties are hardcoded names like `Buyer` or `Seller`) to a production-grade model driven by **Okta (IdP)** and **Daml User Management**.

---

## 1. The Identity Mapping (Okta to Daml)

We establish a "Identity Bridge" in the Go backend that translates external authentication into ledger authorization.

### Data Flow:
1.  **Authentication:** User logs into the Astro frontend via Okta OIDC.
2.  **Credential:** Okta issues a JWT containing a `sub` (unique subject ID) and `email`.
3.  **Authorization:** The Go backend receives the JWT and uses the `IdentityService` to resolve the `sub` to a **Daml Party ID**.
4.  **Persistence:** The mapping between `Okta Sub` <-> `Daml UserID` <-> `Daml PartyID` is stored in the `user_config` database.

| Identity Layer | Example Value |
| :--- | :--- |
| **Okta Subject** | `00u123456789abc` |
| **Daml UserID** | `user_00u123456789abc` |
| **Daml PartyID** | `Alice_Contractor::1220abc...` |

---

## 2. Just-In-Time (JIT) Provisioning

When a new user authenticates via Okta for the first time:
1.  **Party Allocation:** The Go backend calls the Daml JSON API `/v2/parties/allocate` to create a new unique Party on the ledger.
2.  **User Creation:** The backend calls `/v2/users/create` to establish a new Daml User.
3.  **Rights Assignment:** The User is granted `CanActAs(PartyID)` and `CanReadAs(PartyID)`.
4.  **Registry:** The `user_config` database is updated with the new mapping.

---

## 3. Decentralized Invitations (Invite-to-Contract)

To allow existing users to "engage" others (e.g. Buyer inviting a specific Seller by email), we use a ledger-backed invitation pattern.

### The Invitation Lifecycle:
1.  **Creation:** A Buyer drafts a contract and provides the Seller's email.
2.  **EscrowInvitation:** The system creates a Daml `EscrowInvitation` contract.
    *   **Signatory:** Buyer, Mediator.
    *   **Identifier:** A unique `InviteToken` (HMAC-signed hash of email + contract metadata).
    *   **Observer:** The `Unclaimed` public party.
3.  **Hyperlink:** The system generates a URL: `https://escrow.io/onboard?token=TOKEN_ABC`.
4.  **Claiming:**
    *   The recipient clicks the link and authenticates via Okta.
    *   The Go backend verifies the token matches the authenticated email.
    *   The backend exercises the `Claim` choice on the `EscrowInvitation` using the recipient's new JIT-provisioned Party ID.
    *   The `EscrowInvitation` is archived, and a real `EscrowProposal` is created.

---

## 4. Execution Elements (Phase 5)

### Backend (Go)
- **IdentityService:** New service layer for JIT logic and Postgres mapping.
- **JWT Middleware:** Middleware to verify Okta signatures and inject `Daml-User-ID` into request contexts.
- **Invitation API:** `POST /api/v1/invites` to generate tokens and contracts.

### Frontend (Astro)
- **Auth Guard:** Protected routes requiring Okta session.
- **Onboarding Page:** Dynamic landing page (`/onboard`) that parses invitation tokens.
- **Profile Component:** Visual display of the user's Daml Party ID and Wallet status.

### Contracts (Daml)
- **Template `EscrowInvitation`:**
    ```daml
    template EscrowInvitation
      with
        inviter : Party
        mediator : Party
        inviteeEmail : Text
        tokenHash : Text
        terms : EscrowTerms
      where
        signatory inviter, mediator
        choice Claim : ContractId EscrowProposal
          with
            claimingParty : Party
          controller claimingParty
          do
            -- Logic to convert invitation to proposal
    ```
