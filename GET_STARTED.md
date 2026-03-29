# Getting Started with the Stablecoin Escrow Platform

Follow these steps to set up the high-assurance development environment.

## 1. Installation

### Clone & Core Dependencies
```bash
git clone https://github.com/dhushon/daml-escrow.git
cd daml-escrow
```
*   **Go 1.24+**: Required for the backend API.
*   **Java 17**: Required for the Canton Ledger.
*   **DPM**: [Daml Package Manager](https://docs.digitalasset.com/build/3.4/dpm/dpm.html) for contract builds.

### External API Configuration (Action Required)
The platform utilizes several external services. Create a `.env` file in the root:
```bash
# Auth0 / Okta Configuration
AUTH_ISSUER=https://<your-tenant>.auth0.com/
AUTH_CLIENT_ID=<your-id>
AUTH_AUDIENCE=https://escrow-api.com

# Noves Analytics (Task 6.3)
# NOTE: Currently defaults to High-Assurance Simulation mode.
# To enable real-time on-ledger verification, provide your API key below:
NOVES_API_KEY=<your-noves-key> 

# Stablecoin Oracle
ORACLE_WEBHOOK_SECRET=test-secret-123
```

### Contract Compilation
```bash
cd contracts
dpm build --all
cd ..
```

------------------------------------------------------------------------

## 2. Run & Test

### Launch the Distributed Ledger (3-Node)
This starts separate Bank, Buyer, and Seller nodes with isolated PostgreSQL persistence.
```bash
docker-compose -f docker-compose.distributed.yml up -d
```

### Synchronize Package IDs
Ensures the Go API is aware of the latest contract hashes.
```bash
make sync
```

### Start the Backend API
```bash
# Enabled LEDGER_VERBOSE=true to see Noves-ready discovery logs
LEDGER_VERBOSE=true go run cmd/escrow-api/main.go
```

### Start the Frontend
```bash
cd frontend
npm install
npm run dev
```
Navigate to `http://localhost:4321` to access the dashboard.

------------------------------------------------------------------------

## 3. Makefile Commands

The root `Makefile` provides several utility targets:

*   `make test`: Run all Go unit tests.
*   `make integration-test`: Run local integration tests against a single-node sandbox.
*   `make distributed-test`: Run multi-node distributed tests (requires the 3-node docker topology).
*   `make sync`: Update `ledger-state.json` with current package IDs from the ledger.
*   `make up`: (Legacy) Starts simple single-node sandbox.
