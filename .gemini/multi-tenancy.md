# Multi-Tenancy & Data Privacy Strategy (Daml 3.x)

## 1. Core Philosophy: Logical Isolation
We leverage the native **Daml User Management Service (UMS)** to provide strict logical isolation on a shared synchronizer. This allows multiple companies and residential users to co-exist while ensuring data is only visible to authorized parties.

> "The User Management Service enables the decoupling of application users from ledger parties... making it easier to manage permissions and integrate with external Identity Providers." — [Daml 3.x User Management](https://docs.daml.com/deploy/user-management.html)

## 2. Identity Mapping (User-to-Parties)
We employ a **1:N Mapping Model**:
- **Daml User:** A stable, human-readable handle (e.g., `u-google-123`) derived from the Google OIDC `sub`.
- **Daml Parties:** Opaque cryptographic identifiers.
    - **Residential:** Each user owns one primary party.
    - **Corporate:** Shared parties representing organizations (e.g., `DataCloud::1220...`).
- **Rights:** Users are granted `actAs` (write) or `readAs` (read) rights for specific parties via the UMS. 

## 3. Session-Aware Architecture (The "Context" Rule)
The backend MUST be **stateless and session-aware**. Every interaction with the ledger is scoped to the requesting user.

- **Middleware:** Extracts the OIDC token and injects the `AuthSubKey` (Daml User ID) into the Go context.
- **Service Layer:** All methods must accept a `userID` parameter.
- **Submission Rule:** Every JSON API call MUST include the `userId` in the payload.
> "When a command is submitted with a `userId`, the ledger automatically validates that the user has the required `actAs` or `readAs` rights for the parties involved." — [Daml JSON API v2](https://docs.daml.com/json-api/index.html)

## 4. Privacy & Visibility Guardrails
- **Ledger-Enforced Privacy:** In Daml, a party can ONLY see a contract if they are a **Signatory** or an **Observer**. The backend must never attempt to bypass this by using high-privilege "admin" parties for user-scoped queries.
- **On-Behalf-Of Operations:** Services acting "on behalf of" an organization (e.g., automated settlement) MUST use a specific User ID that has been granted limited rights for the corporate party.
- **Organizational Context:** Use the `X-Org-Context` header or email domain to dynamically switch between a user's residential and corporate viewing scopes.

## 5. "Invite Participant" Affiliation Lock
To prevent identity spoofing during onboarding:
1. Invitations are issued to an **Email Address**.
2. **Domain Extraction:** If the email belongs to a verified business domain (e.g., `@datacloud.com`), the invitation is logically associated with that organization.
3. **Claim Constraint:** Only a user whose authenticated OIDC `email` claim matches the invitation (and domain) can execute the `Claim` choice.

## 6. Official Documentation References
- [User Management Service Overview](https://docs.daml.com/concepts/user-management.html)
- [Multi-Tenancy Patterns in Canton](https://docs.digitalasset.com/canton/architecture/multi-tenancy.html)
- [JSON API V2 Specification](https://docs.daml.com/json-api/v2-spec.html)
