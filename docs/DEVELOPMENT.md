# Tripartite Development Guide

This document authoritatively defines the local development model for the high-assurance Stablecoin Escrow platform.

## 1. Orchestration Model: What Runs Where

The platform authoritatively mirrors the tripartite GKE topology in the local environment using **Docker Compose**.

| Component | Container Name | Logical Perimeter | Default Port |
| :--- | :--- | :--- | :--- |
| **Unified Gateway** | `escrow-gateway-dev` | Ingress Simulator | `8080` |
| **Bank API** | `bank-api-dev` | Bank Sovereign Node | `8081` (internal) |
| **Buyer API** | `buyer-api-dev` | Buyer Sovereign Node | `8081` (internal) |
| **Seller API** | `seller-api-dev` | Seller Sovereign Node | `8081` (internal) |
| **Canton Ledger** | `escrow-ledger-dev` | Distributed Ledger | `7575-7577` |
| **Postgres DB** | `escrow-postgres-dev` | Shared Multi-DB | `5432` |

## 2. Institutional Service Endpoints

Developers MUST authoritatively route all tripartite traffic through the **Unified Gateway (8080)** to maintain environment parity.

*   **Bank Services**: `http://localhost:8080/bank/api/v1/*`
*   **Buyer Services**: `http://localhost:8080/buyer/api/v1/*`
*   **Seller Services**: `http://localhost:8080/seller/api/v1/*`
*   **Global Health**: `http://localhost:8080/health` (Gateway check)

## 3. High-Assurance Observability

| Tool | URL | Role |
| :--- | :--- | :--- |
| **Jaeger** | `http://localhost:16686` | Tripartite distributed tracing. |
| **Grafana** | `http://localhost:3000` | Real-time performance dashboards. |
| **Prometheus** | `http://localhost:9090` | Metrics time-series engine. |

## 4. Local Development Workflow

### A. Initialize the Distributed Ledger
```bash
docker-compose -f docker-compose.distributed.yml up -d
make sync
```

### B. Launch the Observability Stack
```bash
docker-compose -f docker-compose.otel.yml up -d
```

### C. Execute Tripartite Verification
```bash
# Verify Bank API via Gateway
curl http://localhost:8080/bank/api/v1/health
```

## 5. Identity & Authorization Simulation

To simulate institutional roles without real OIDC tokens:
1.  Run the API with the `--bypass` flag.
2.  Pass the **`X-Dev-User`** header with an email from the `config/identity_providers.yaml` (e.g., `joey@buyer.com`).

---
**High-Assurance Standard**: All local development MUST authoritatively synchronize with this guide to ensure GKE production readiness.
