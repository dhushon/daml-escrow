# Daml & Canton Ledger Requirements (SDK 3.4.x)

## 1. Authorization & Identity (Phase 5+)
- **Party ID Integrity:** Command submissions MUST use fully qualified Party IDs (e.g., `Buyer::1220abc...`).
- **User ID Scoping:** The `userId` MUST match the sanitized Daml User ID derived from the external IdP (e.g., `u-google-sub-123`).
- **JIT Provisioning Flow (SDK 3.4.x specific):** 
    1. **Allocation:** `POST /v2/parties` with `{ "partyIdHint": "id", "displayName": "..." }`.
    2. **User Creation:** `POST /v2/users` with a nested user object:
       ```json
       {
         "user": {
           "id": "u-id",
           "primaryParty": "party-id",
           "isDeactivated": false,
           "identityProviderId": ""
         }
       }
       ```
    3. **Rights Granting:** `POST /v2/users/{id}/rights` requires `userId` and `identityProviderId` in the body along with the `grant` array.
- **Dynamic Resolution:** NEVER hardcode Party IDs. Always resolve via the `partyMap` cache refreshed from `/v2/parties`.

## 2. JSON API V2 Serialization
- **Nullary Constructors:** Data constructors with no fields MUST be represented as a plain string: `"payload": "ApproveMilestoneArg"`.
- **Zero-Argument Choices:** Choices with no parameters MUST use an empty object: `"choiceArgument": {}`.
- **Interface Exercises:** Targeting an interface choice MUST use the Interface Package ID and Template ID in the `ExerciseCommand`.

## 3. Response Handling
- **NDJSON Support:** `/v2/state/active-contracts` returns Newline Delimited JSON. Parsers MUST handle streaming responses.
- **Event Extraction:** Check both `CreatedEvent` and `createdEvent` when parsing transaction trees.

## 4. Stability & Performance
- **Timeout Management:** Ledger transactions involving automation or complex retry logic REQUIRE a minimum 120s client timeout.
- **Retry Strategy:** `GetEscrow` operations should use a minimum of 15 retries with 2s delays to account for indexing latency in multi-participant Canton environments.
