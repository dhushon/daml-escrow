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

## DAR Update & Ledger Reset Process
When modifying contract logic (Daml source) or adding new templates, the ledger environment MUST be synchronized:

1.  **Build All:** Execute `make daml-build` (uses `dpm build --all`) to ensure all interfaces, implementations, and test packages are compiled and Package IDs are updated.
2.  **Verify Backend IDs:** Inspect the built implementation DAR using `~/.dpm/bin/dpm inspect-dar <file>` to extract the new Package ID. Update `PackageID` and `InterfacePackageID` in `internal/ledger/json_base.go`.
3.  **Wipe State:** Execute `docker-compose down -v` to definitively clear persistent database schemas and ledger state.
4.  **Re-upload & Bootstrap:** Execute `make up` to restart the stack. This triggers the `db-init` container and the `escrow-ledger` bootstrap script (`sandbox_init.canton`), ensuring fresh DARs are uploaded and topology is re-mapped.
5.  **Integration Pass:** Run `make integration-test` to verify that the backend client correctly interacts with the new contract logic.
