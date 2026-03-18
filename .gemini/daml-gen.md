# DAML & Go-DAML Integration Rules

## Critical Version Compatibility
- **Ledger API Version:** The backend Go services (via `go-daml`) are currently built for **Ledger API v2**.
- **SDK Requirement:** **Daml SDK 3.0 or higher** is MANDATORY to support Ledger API v2.
- **Legacy Conflict:** Daml SDK 2.10.x and earlier only support Ledger API v1. Using SDK 2.x with the current Go backend will result in `rpc error: code = Unimplemented desc = Method not found: com.daml.ledger.api.v2.CommandService/SubmitAndWait`.

## Java Environment
- **Requirement:** **JDK 17** (LTS).
- **Rationale:** Required for Daml 2.10/3.0+ CLI tools and Canton components. Newer versions (like JDK 25) cause connection timeouts.

## Tooling
- **SDK 3.0+ Management:** Use **`dpm`** (Digital Asset Package Manager) instead of the legacy `daml` assistant.
- **Installation:** `curl https://get.digitalasset.com/install/install.sh | sh`

## DAR & DALF Handling
- **Build Command:** `cd contracts && dpm build` (or `daml build` if on 2.x).
- **Package ID Extraction:** Use `unzip -p <dar_path> META-INF/MANIFEST.MF` to find the `Main-Dalf` entry. The hex string is the Package ID.

## Manual Go Binding Pattern (Fallback)
When automated codegen is unavailable (as it currently targets LF 2.x/SDK 3.0), implement bindings in `internal/ledger/generated/` following the established pattern in `stablecoin_escrow.go`.

## Ledger Connection
- **Host:** Prefer `127.0.0.1` over `localhost` in `config.yaml` to avoid IPv6 (`[::1]`) connection refused errors on some systems.
- **Client:** `github.com/smartcontractkit/go-daml/pkg/client.DamlBindingClient`.
- **Submission:** Always use `model.SubmitAndWaitRequest` with a unique `CommandID`.
