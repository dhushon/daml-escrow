# Tripartite Development Guide

This document authoritatively defines the local development model for the high-assurance Stablecoin Escrow platform.

## 1. Orchestration Model: What Runs Where

The platform authoritatively mirrors the tripartite GKE topology in the local environment using **Docker Compose**.

| Component | Container Name | Logical Perimeter | Default Port |
| :--- | :--- | :--- | :--- |
| **Unified Gateway** | `escrow-gateway-dev` | Ingress Simulator | `8080` |
| **Bank API** | `bank-api-dev` | Bank Sovereign Node | `8081` (internal) |
| **Depositor API** | `depositor-api-dev` | Depositor Sovereign Node | `8081` (internal) |
| **Beneficiary API** | `beneficiary-api-dev` | Beneficiary Sovereign Node | `8081` (internal) |
| **Canton Ledger** | `escrow-ledger-dev` | Distributed Ledger | `7575-7577` |
| **Postgres DB** | `escrow-postgres-dev` | Shared Multi-DB | `5432` |

## 2. Institutional Service Endpoints

Developers MUST authoritatively route all tripartite traffic through the **Unified Gateway (8080)** to maintain environment parity.

*   **Bank Services**: `http://localhost:8080/bank/api/v1/*`
*   **Depositor Services**: `http://localhost:8080/depositor/api/v1/*`
*   **Beneficiary Services**: `http://localhost:8080/beneficiary/api/v1/*`
*   **Global Health**: `http://localhost:8080/health` (Gateway check)

## 3. High-Assurance Observability

| Tool | URL | Role |
| :--- | :--- | :--- |
| **Jaeger** | `http://localhost:16686` | Distributed tracing. |
| **Grafana** | `http://localhost:3000` | Real-time performance dashboards. |
| **Prometheus** | `http://localhost:9090` | Metrics time-series engine. |

## 4. Local Development Workflow

The platform supports three development tiers via the root `Makefile`.

### A. Standalone (Single-Node)
Best for core UX and API development.
```bash
make standalone-up
```

### B. Tripartite (Distributed)
Required for testing cross-node privacy and distributed synchronization.
```bash
make tri-up
```

### C. GCP Proxy (Hybrid)
Points local services to a GKE-hosted ledger environment.
```bash
make pilot-local
```

## 5. Identity & Authorization Simulation

To simulate institutional roles without real OIDC tokens:
1.  Run the API with the `--bypass` flag (enabled by default in `make *-up` targets).
2.  Pass the **`X-Dev-User`** header with an email from the `config/identity_providers.yaml` (e.g., `joey@depositor.com`).

---
**High-Assurance Standard**: All local development MUST authoritatively synchronize with this guide to ensure GKE production readiness.
