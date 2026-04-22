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
The platform utilizes several external services. 

👉 **[Mandatory: Setup Okta OIDC Identity Provider](./IDENTITY.md)**

1.  **Terraform Provisioning:** The identity infrastructure is managed via Terraform.
    ```bash
    cd terraform
    cp terraform.tfvars.example terraform.tfvars
    # Provide your Okta Org and API Token
    terraform init && terraform apply
    ```
2.  **Automated Personas:** Create the demonstration users (Joey, Jimmy, etc.) with a single script:
    ```bash
    ./scripts/setup_test_users.sh
    ```

### Institutional Custody Setup (Optional)
To use real stablecoin movements instead of the default mock, configure one of the following institutional providers:

**BitGo Enterprise:**
*   Requires a local [BitGo Express](https://github.com/BitGo/bitgo-express) proxy.
*   `BITGO_EXPRESS_URL`: URL of your local proxy (e.g., `http://localhost:3080`).
*   `BITGO_ACCESS_TOKEN`: Your BitGo API V2 token.
*   `BITGO_COIN`: Coin identifier (e.g., `teth:usdc`).

**Circle WaaS:**
*   `CIRCLE_BASE_URL`: API endpoint (defaults to `https://api.circle.com`).
*   `CIRCLE_API_KEY`: Your Circle API Key.
*   `CIRCLE_ENTITY_SECRET`: Your Entity Secret for developer-controlled wallets.

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
The API uses a Cobra-based CLI. Use the `--bypass` flag for local development without real OIDC tokens.
```bash
# Start in Dev mode with Auth Bypass enabled
go run cmd/escrow-api/main.go serve --env dev --bypass
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
