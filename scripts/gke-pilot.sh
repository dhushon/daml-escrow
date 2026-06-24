#!/bin/bash
# scripts/gke-pilot.sh --- Tier 2: Workload Orchestration
# authoritatively manages GKE cluster, tripartite nodes, and pilot deployments.

set -e

# Find project root directory relative to this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load environment variables from project root .env
if [ -f "$ROOT_DIR/.env" ]; then
    export $(grep -v '^#' "$ROOT_DIR/.env" | grep -v '^$' | xargs)
fi

# Set dynamic variables based on environment variables or defaults
export TF_VAR_project_id="${TF_VAR_project_id:-vdcai-daml}"
export TF_VAR_region="${TF_VAR_region:-us-central1}"
export TF_VAR_environment="${TF_VAR_environment:-dev}"

PROJECT_ID="$TF_VAR_project_id"
REGION="$TF_VAR_region"
ZONE="${REGION}-a"
CLUSTER_NAME="escrow-cluster-${TF_VAR_environment}"


function log() {
  echo "------------------------------------------------------------------------"
  echo "$1"
  echo "------------------------------------------------------------------------"
}

function up() {
  log "TIER 2: PROVISIONING WORKLOAD INFRASTRUCTURE"
  cd terraform/workload
  terraform init
  terraform apply -auto-approve
  cd ../..

  log "VERIFYING AUTHORITATIVE CONTEXT: $CLUSTER_NAME"
  gcloud container clusters get-credentials $CLUSTER_NAME --zone $ZONE --project $PROJECT_ID

  log "[1/5] INITIALIZING TRI-PARTITE BOUNDARIES"
  kubectl apply -f k8s/namespaces.yaml

  log "[2/5] DEPLOYING IDENTITY CONTROLLERS"
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
  # Patch Ingress to use our reserved Regional Static IP (from Tier 1)
  kubectl patch service ingress-nginx-controller -n ingress-nginx -p '{"spec": {"loadBalancerIP": "34.31.124.124"}}'

  log "BRING-UP COMPLETE: Awaiting GKE LoadBalancer & Certificate propagation."
}

function down() {
  log "TIER 2: DESTROYING EXPENSIVE WORKLOAD INFRASTRUCTURE"
  
  # [1/2] Purge K8s resources to ensure clean load balancer detachment
  kubectl delete -f k8s/ingress.yaml --ignore-not-found || true
  kubectl delete namespaces bank depositor beneficiary --ignore-not-found || true
  
  # [2/2] Authoritative Terraform destruction
  cd terraform/workload
  terraform init
  terraform destroy -auto-approve
  cd ../..
  
  log "TAKEDOWN COMPLETE: GKE Cluster and node pools definitively purged."
}

function status() {
  log "AUDITING PILOT HEALTH"
  kubectl get pods --all-namespaces | grep -E "bank|buyer|seller"
  kubectl get ingress -n bank
  kubectl get certificate --all-namespaces
}

case "$1" in
  up) up ;;
  down) down ;;
  status) status ;;
  *) echo "Usage: $0 {up|down|status}" ;;
esac
|status}" ;;
esac
