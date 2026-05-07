# GKE Sovereign Orchestration Guide

This directory contains the high-assurance Kubernetes manifests for deploying the Stablecoin Escrow platform in a participant-sovereign configuration.

## Deployment Topology
Each participant (**Bank**, **Buyer**, **Seller**) is deployed into a dedicated, isolated Kubernetes namespace.

- **Canton Nodes**: Deployed as `StatefulSets` to ensure persistent identity and high-availability.
- **Go APIs**: Deployed as isolated `Deployments` authoritatively locked to their respective Canton node via `PARTICIPANT_ID`.

## Release Engineering (Artifact Registry)
For production pilots, images must be authoritatively pushed to the high-assurance GCP repository.

1. **Authenticate Docker**:
   ```bash
   gcloud auth configure-docker us-central1-docker.pkg.dev --project=vdcai-daml --quiet
   ```

2. **Push Sovereign Image**:
   ```bash
   docker build -t us-central1-docker.pkg.dev/vdcai-daml/escrow-platform-dev/escrow-api:latest .
   docker push us-central1-docker.pkg.dev/vdcai-daml/escrow-platform-dev/escrow-api:latest
   ```

## Deployment Sequence

### 1. Initialize Boundaries
Establish the logical perimeters for each participant:
```bash
kubectl apply -f k8s/namespaces.yaml
```

### 2. High-Assurance Identity (TLS & mTLS)
Apply the authoritative trust anchors and issuers:
```bash
# Public Let's Encrypt Identity
kubectl apply -f k8s/tls-issuer.yaml

# Internal GCP CAS Identity (mTLS)
kubectl apply -f k8s/cas-issuer.yaml
```

### 3. Hardened Entrypoint (Ingress)
Apply the unified gateway with path-based routing:
```bash
kubectl apply -f k8s/ingress.yaml
```

### 4. Deploy Tripartite sovereign Stacks
```bash
# Bank
kubectl apply -f k8s/bank-ledger.yaml
kubectl apply -f k8s/bank-api.yaml

# Buyer & Seller
kubectl apply -f k8s/buyer-ledger.yaml
kubectl apply -f k8s/buyer-api.yaml
kubectl apply -f k8s/seller-ledger.yaml
kubectl apply -f k8s/seller-api.yaml
```

### 5. Pilot Verification (vdatacloudai.com)
Verify that the domain is authoritatively mapped and protected by TLS 1.3:
```bash
curl -v https://api.vdatacloudai.com/bank/api/v1/health
```

## Security & Secrets
Sensitive credentials (Okta Client Secret, Stablecoin Tokens) are authoritatively vended from **GCP Secret Manager** at runtime. Ensure the GKE Node Pool has the necessary IAM permissions to access these secrets.
