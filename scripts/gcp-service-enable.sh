#!/bin/bash
# High-Assurance GCP Service Enablement Script
# Authoritatively enables mandatory APIs and synchronizes cross-service identity.

PROJECT_ID="vdcai-daml"

echo "------------------------------------------------------------------------"
echo "ENABLING MANDATORY GCP SERVICES FOR PROJECT: $PROJECT_ID"
echo "------------------------------------------------------------------------"

SERVICES=(
  "secretmanager.googleapis.com"
  "artifactregistry.googleapis.com"
  "compute.googleapis.com"
  "container.googleapis.com"
  "logging.googleapis.com"
  "monitoring.googleapis.com"
  "dns.googleapis.com"
  "privateca.googleapis.com"
  "cloudkms.googleapis.com"
)

for service in "${SERVICES[@]}"; do
  echo "Enabling $service..."
  gcloud services enable $service --project=$PROJECT_ID --quiet
done

# --- Cross-Service Identity Synchronization ---

echo "------------------------------------------------------------------------"
echo "SYNCHRONIZING GKE IDENTITY FOR ARTIFACT REGISTRY ACCESS"
echo "------------------------------------------------------------------------"

PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format="value(projectNumber)")
GKE_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

echo "Authorizing Service Account: $GKE_SA"
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$GKE_SA" \
  --role="roles/artifactregistry.reader" --quiet > /dev/null

echo "------------------------------------------------------------------------"
echo "SUCCESS: High-Assurance Services & Identity Synchronized."
echo "------------------------------------------------------------------------"
