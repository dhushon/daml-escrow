# Canton Ledger Management Guide (Daml 3.4.11)

This guide documents the setup, configuration, and integration of the Canton 3.x ledger for the Stablecoin Escrow project.

## 1. Critical Lessons Learned (Stabilization Phase)

### Daml Language Version
- **Constraint:** Canton 3.x requires Daml LF **2.1** or higher.
- **Fix:** `contracts/daml.yaml` must include:
  ```yaml
  build-options:
    - --target=2.1
  ```
- **Error:** `ALLOWED_LANGUAGE_VERSIONS` or `Expected version between 2.1 and 2.2 but got 1.14`.

### Bootstrap & Initialization (`sandbox_init.canton`)
- **Node References:** Use generic positional access (`participants.all.head`) rather than variable names like `local` or `sandbox` to avoid "not found" errors in the console.
- **Manual Setup:** Even in a single-node setup, the synchronizer must be explicitly bootstrapped in `daemon` mode:
  ```scala
  bootstrap.synchronizer_local("mysynchronizer")
  ```
- **Persistence:** Never include `sys.exit(0)` at the end of a bootstrap script passed via `--bootstrap` to a daemon, or the container will terminate immediately upon completion.

### Docker Networking & Health
- **Health Checks:** Use low-level `/proc/net/tcp` lookups to avoid dependencies like `nc` or `curl` in minimal base images:
  ```yaml
  test: ["CMD-SHELL", "grep -q '1D97' /proc/net/tcp || exit 1"] # 1D97 is port 7575
  ```
- **Postgres Persistence:**
  - Nodes (Participant, Sequencer, Mediator) **cannot** share the same database schema.
  - In our setup, the **Participant** uses Postgres, while **Sequencer/Mediator** use Memory to avoid schema conflicts while maintaining user/party persistence.

---

## 2. Sandbox Operations

### Automated Local Setup (Host)
This is the recommended way to run the ledger for development. It uses the `dpm sandbox` wrapper which handles internal connections automatically.

```bash
# Clean, Build, Start, and Setup Topology/Users
make sandbox-up
```

### Automated Docker Setup
This runs the full stack (Postgres, Ledger, API) in containers. It uses the `daemon` mode with explicit node definitions.

```bash
# Deploy full persistent stack
make docker-up

# Verify health
docker-compose ps
```

### Manual Configuration Management
- **Local Config:** `contracts/Sandbox/sandbox.conf`
- **Docker Config:** `contracts/Sandbox/sandbox-docker.conf`
- **Init Script:** `contracts/Sandbox/sandbox_init.canton`

---

## 3. JSON API V2 Integration

### User Management
User creation and rights management must follow the strict V2 schema:

**Create User:**
```bash
curl -X POST http://localhost:7575/v2/users \
  -d '{"user": {"id": "Buyer", "primaryParty": "PARTY_ID", "isDeactivated": false, "identityProviderId": ""}}'
```

**Grant Rights:**
```bash
curl -X POST http://localhost:7575/v2/users/Buyer/rights \
  -d '{"userId": "Buyer", "actAs": ["PARTY_ID"], "identityProviderId": ""}'
```

## 4. Verification
After any ledger restart, run the integration tests to confirm full lifecycle functionality:
```bash
make integration-test
```
