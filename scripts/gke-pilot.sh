#!/bin/bash
# High-Assurance GKE Pilot Management Script
# authoritatively manages the bring-up and clean-up of the Stablecoin Escrow platform.

PROJECT_ID="vdcai-daml"
REGION="us-central1"
ZONE="us-central1-a"
CLUSTER_NAME="escrow-cluster-dev"

function log() {
  echo "------------------------------------------------------------------------"
  echo "$1"
  echo "------------------------------------------------------------------------"
}

function up() {
  log "VERIFYING AUTHORITATIVE CONTEXT: $CLUSTER_NAME"
  gcloud container clusters get-credentials $CLUSTER_NAME --zone $ZONE --project $PROJECT_ID

  log "[1/5] INITIALIZING TRI-PARTITE BOUNDARIES"
  kubectl apply -f k8s/namespaces.yaml

  log "[2/5] DEPLOYING IDENTITY CONTROLLERS"
  # Core TLS/Identity infrastructure
  kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.4/cert-manager.yaml
  kubectl apply -f https://github.com/jetstack/google-cas-issuer/releases/download/v0.8.0/google-cas-issuer-v0.8.0.yaml
  kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml

  log "[3/5] ESTABLISHING HIGH-ASSURANCE TRUST ANCHORS"
  kubectl apply -f k8s/tls-issuer.yaml
  kubectl apply -f k8s/cas-issuer.yaml

  log "[4/5] PROVISIONING SOVEREIGN WORKLOADS"
  kubectl apply -f k8s/canton-configs.yaml
  for ns in bank buyer seller; do
    kubectl create secret generic db-secrets --namespace $ns --from-literal=password=escrow --dry-run=client -o yaml | kubectl apply -f -
  done
  kubectl apply -f k8s/bank-ledger.yaml && kubectl apply -f k8s/bank-api.yaml
  kubectl apply -f k8s/buyer-ledger.yaml && kubectl apply -f k8s/buyer-api.yaml
  kubectl apply -f k8s/seller-ledger.yaml && kubectl apply -f k8s/seller-api.yaml

  log "[5/5] HARDENING PUBLIC ENTRYPOINT"
  kubectl apply -f k8s/ingress.yaml
  # Patch Ingress to use our reserved Regional Static IP
  kubectl patch service ingress-nginx-controller -n ingress-nginx -p '{"spec": {"loadBalancerIP": "34.31.124.124"}}'

  log "BRING-UP COMPLETE: Awaiting GKE LoadBalancer & Certificate propagation."
}

function down() {
  log "PURGING TRI-PARTITE PILOT STACK"
  kubectl delete -f k8s/ingress.yaml --ignore-not-found
  kubectl delete namespaces bank buyer seller --ignore-not-found
  kubectl delete namespace observability --ignore-not-found
  log "CLEAN-UP COMPLETE."
}

function status() {
  log "AUDITING PILOT HEALTH"
  kubectl get pods --all-namespaces -l "env=dev"
  kubectl get ingress -n bank
  kubectl get certificate --all-namespaces
}

case "$1" in
  up) up ;;
  down) down ;;
  status) status ;;
  *) echo "Usage: $0 {up|down|status}" ;;
esac
