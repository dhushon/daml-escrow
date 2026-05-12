#!/bin/bash
# scripts/admin-setup.sh --- Tier 1: Administrative Foundation
# authoritatively manages GCP Root CA, Audit Sinks, and DNS.

set -e

PROJECT_ID=${1:-vdcai-daml}
ACTION=${2:-status}

echo "------------------------------------------------------------------------"
echo "TIER 1: ADMINISTRATIVE FOUNDATION --- ACTION: $ACTION"
echo "------------------------------------------------------------------------"

if [ "$ACTION" == "up" ]; then
    echo "Provisioning Tier 1 infrastructure (CAS, Audit, IP)..."
    cd terraform/admin
    terraform init
    terraform apply -auto-approve -var="project_id=$PROJECT_ID"
    
    echo "Finalizing CAS IAM bindings..."
    # High-Assurance: Bind both Service Account AND the CORRECT Workload Identity member
    gcloud privateca pools add-iam-policy-binding escrow-ca-pool-dev \
        --location=us-central1 \
        --member="serviceAccount:cert-manager-google-cas-issuer@$PROJECT_ID.iam.gserviceaccount.com" \
        --role="roles/privateca.certificateRequester" \
        --project=$PROJECT_ID --quiet || true

    gcloud privateca pools add-iam-policy-binding escrow-ca-pool-dev \
        --location=us-central1 \
        --member="serviceAccount:$PROJECT_ID.svc.id.goog[cert-manager/cert-manager-google-cas-issuer]" \
        --role="roles/privateca.certificateRequester" \
        --project=$PROJECT_ID --quiet || true

elif [ "$ACTION" == "status" ]; then
    echo "Auditing Tier 1 Health..."
    gcloud privateca pools list --location=us-central1 --project=$PROJECT_ID
    gcloud compute addresses list --filter="name:escrow-gateway-ip-dev" --project=$PROJECT_ID
    gcloud logging sinks list --project=$PROJECT_ID | grep "escrow-compliance"

else
    echo "Unknown action: $ACTION. Supported: up, status"
    exit 1
fi
