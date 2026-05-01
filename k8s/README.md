# GKE Sovereign Orchestration Guide

This directory contains the high-assurance Kubernetes manifests for deploying the Stablecoin Escrow platform in a participant-sovereign configuration.

## Deployment Topology
Each participant (**Bank**, **Buyer**, **Seller**) is deployed into a dedicated, isolated Kubernetes namespace.

- **Canton Nodes**: Deployed as `StatefulSets` to ensure persistent identity and high-availability.
- **Go APIs**: Deployed as isolated `Deployments` authoritatively locked to their respective Canton node via `PARTICIPANT_ID`.

## Deployment Sequence

### 1. Initialize Boundaries
Establish the logical perimeters for each participant:
```bash
kubectl apply -f k8s/namespaces.yaml
```

### 2. Provision Storage & Configuration (TBD)
Apply the necessary `PersistentVolumeClaims` and `ConfigMaps` for the Canton nodes.

### 3. Deploy Bank Sovereign Stack
```bash
kubectl apply -f k8s/bank-ledger.yaml
kubectl apply -f k8s/bank-api.yaml
```

### 4. High-Assurance Verification
Verify that the Bank's API can only communicate with the Bank's ledger:
```bash
kubectl get pods -n bank
kubectl logs -f deployment/bank-api -n bank
```

## Security & Secrets
Sensitive credentials (Okta Client Secret, Stablecoin Tokens) are authoritatively vended from **GCP Secret Manager** at runtime. Ensure the GKE Node Pool has the necessary IAM permissions to access these secrets.
