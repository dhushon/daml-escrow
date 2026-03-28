# Canton Ledger Management Guide (Daml 3.4.11)

This guide documents the setup, configuration, and integration of the Canton 3.x ledger for the Stablecoin Escrow project.

## 1. Canton Infrastructure & Core APIs

The escrow service utilizes the Canton Network's institutional-grade features for confidentiality and interoperability.

### Daml SDK (Canton Ledger API)
The primary interface for defining the contractual state of an escrow. This project uses the **Daml JSON API V2** to interact with contracts on Canton participant nodes.

### CIP-0056 Token Standard
All stablecoin assets (e.g., **USDCx via BitGo/Circle**) adhere to the **CIP-0056** standard.
- **Holding Interface:** Manages the custody of escrowed assets.
- **Transfer Interface:** Ensures secure, interoperable movement across applications on the Canton Network.
- **Implementation:** Leverages the `simple-token` package which defines the 6 core CIP-0056 on-ledger interfaces.

### OpenZeppelin Stablecoin/CDP Module
We utilize Daml templates from the **Canton Stablecoin** repository for:
- **CDP-style Vaults:** Collateralized Debt Positions for overcollateralized stablecoin management.
- **Minting & Liquidation:** Extending the basic token interfaces to handle complex escrow pledging systems.

### Advanced Monitoring & Analytics
- **Splice Validator APIs:** Provides high-level endpoints (e.g., External Signing API) for automated workflows and balance checks for external parties like trusted escrow agents.
- **Noves Data & Analytics API:** Used for real-time and indexed data tracking of token holdings, transaction events, and wallet metrics.

---

## 2. Critical Lessons Learned (Stabilization Phase)

### Daml Language Version
- **Constraint:** Canton 3.x requires Daml LF **2.1** or higher.
- **Fix:** `contracts/daml.yaml` must include:
  ```yaml
  build-options:
    - --target=2.1
  ```

### Bootstrap & Initialization (`sandbox_init.canton`)
- **Node References:** Use generic positional access (`participants.all.head`) rather than variable names like `local` or `sandbox`.
- **Manual Setup:** The synchronizer must be explicitly bootstrapped in `daemon` mode.
- **Persistence:** Never include `sys.exit(0)` at the end of a bootstrap script passed via `--bootstrap` to a daemon.

---

## 3. Sandbox Operations

### Automated Local Setup (Host)
```bash
# Clean, Build, Start, and Setup Topology/Users
make sandbox-up
```

### Automated Docker Setup
```bash
# Deploy full persistent stack
make docker-up
```

---

## 4. JSON API V2 Integration

### User Management
User creation and rights management must follow the strict V2 schema.

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

## 5. Verification
After any ledger restart, run the integration tests to confirm full lifecycle functionality:
```bash
make integration-test
```
