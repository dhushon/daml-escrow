# Daml & Canton Ledger Requirements (SDK 3.4.x)

## Authorization Rules
- **Party ID Integrity:** The `actAs` field in any command submission MUST use the fully qualified Party ID (e.g., `Buyer::1220abc...`).
- **User ID Scoping:** The `userId` field MUST use the plain user name (e.g., `Buyer`) to match the JWT/User mapping.
- **Never hardcode Party IDs:** Always use `refreshPartyMap` to resolve user names to their active ledger Party IDs. The `/v2/parties` response is an object with a `partyDetails` array.

## JSON API V2 Serialization
- **Nullary Constructors:** Data constructors with no fields (e.g., `data ApproveMilestoneArg = ApproveMilestoneArg`) MUST be represented as a plain string in the JSON payload: `"payload": "ApproveMilestoneArg"`.
- **Zero-Argument Choices:** Choices with no parameters (e.g., `choice Settle : ...`) MUST use an empty object for the argument: `"choiceArgument": {}`.
- **Interface ID:** Interface exercises MUST provide both `templateId` (Interface ID for the exercise) and `interfaceId` (Interface ID for validation). In SDK 3.x, the `templateId` in the `ExerciseCommand` should target the Interface.

## Response Handling
- **Flexible JSON Parsing:** The `/v2/state/active-contracts` endpoint may return either a JSON Array or Newline Delimited JSON (NDJSON) depending on the environment. Parsers MUST handle both formats gracefully.
- **Transaction Events:** When extracting contracts from a transaction response, check both `CreatedEvent` and `createdEvent` (case sensitivity varies).
