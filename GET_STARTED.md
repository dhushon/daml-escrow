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

The platform supports three development tiers via the root `Makefile`.

### Standalone (Single-Node)
Best for core UX and API development.
```bash
make standalone-up
```

### Tripartite (Distributed)
Required for testing cross-node privacy and distributed synchronization. Starts separate Bank, Depositor, and Beneficiary nodes.
```bash
make tri-up
```

### GCP Proxy (Hybrid)
Points local services to a GKE-hosted ledger environment.
```bash
make pilot-local
```

Navigate to `http://localhost:4321` to access the dashboard.

------------------------------------------------------------------------

## 3. Makefile Commands

The root `Makefile` provides several utility targets for authoritative orchestration:

*   **`make standalone-up`**: Authoritatively launch local baseline (Single-Node).
*   **`make standalone-down`**: Purge all standalone processes and containers.
*   **`make tri-up`**: Authoritatively launch distributed tripartite stack.
*   **`make tri-down`**: Purge all tripartite processes and containers.
*   **`make bootstrap-local`**: Synchronize DAR packages and allocate Parties on localhost.
*   **`make test`**: Run all Go unit tests.
*   **`make codegen`**: Authoritatively regenerate Go bindings from DAR files.
