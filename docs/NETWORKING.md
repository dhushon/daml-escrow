# high-assurance networking registry

This document authoritatively defines the port mappings and network perimeters for the Stablecoin Escrow platform.

## 1. Public Institutional Interface (GKE)

Traffic is authoritatively routed through the **Unified Gateway** using TLS 1.3.

| Service | Public Port | Internal Path | Scope |
| :--- | :--- | :--- | :--- |
| **Unified Gateway** | `443` | `api.vdatacloudai.com` | Public Entrypoint |
| **Bank Stack** | `443` | `/bank/*` | Sovereign Namespace |
| **Buyer Stack** | `443` | `/buyer/*` | Sovereign Namespace |
| **Seller Stack** | `443` | `/seller/*` | Sovereign Namespace |
| **Health Hub** | `443` | `/bank/api/v1/health` | Recursive Monitoring |

## 2. Local Development & Simulation

The local environment (Docker) perfectly authoritatively mirrors the GKE production topology.

| Service | Port | URL | Role |
| :--- | :--- | :--- | :--- |
| **Unified Gateway** | `8080` | `localhost:8080` | Local Ingress Simulator |
| **Astro Frontend** | `4321` | `localhost:4321` | High-Assurance Dashboard |
| **Canton Ledger** | `7575-7577` | `localhost:7575` | Local Distributed Ledger |

## 3. Observability & Auditing

Dashboards for real-time tripartite monitoring and SOC2-compliant tracing.

| Tool | Port | Purpose | High-Assurance Source |
| :--- | :--- | :--- | :--- |
| **Grafana** | `3000` | Metrics Dashboard | Prometheus |
| **Jaeger** | `16686` | Distributed Tracing | OTEL Collector |
| **Prometheus** | `9090` | Metrics Query Engine | OTEL Collector |
| **OTEL Collector** | `4318` | Telemetry Hub | Participant APIs |

## High-Assurance Networking Rules

1.  **Zero-Trust**: All service-to-service traffic in GKE MUST utilize mTLS on port **443**.
2.  **Explicit Mapping**: Local development MUST authoritatively use the **Unified Gateway (8080)** to maintain tripartite parity.
3.  **Governance**: All port changes MUST be authoritatively documented in this registry.
