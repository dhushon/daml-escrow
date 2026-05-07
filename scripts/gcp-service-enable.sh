#!/bin/bash
# High-Assurance GCP Service Enablement Script
# Authoritatively enables mandatory APIs and synchronizes mTLS identity.

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
  "iam.googleapis.com"
)

for service in "${SERVICES[@]}"; do
  echo "Enabling $service..."
  gcloud services enable $service --project=$PROJECT_ID --quiet
done

# --- mTLS Identity (GCP CAS) Provisioning ---

echo "------------------------------------------------------------------------"
echo "PROVISIONING GCP CAS SERVICE ACCOUNT"
echo "------------------------------------------------------------------------"

SA_NAME="cert-manager-google-cas-issuer"
SA_EMAIL="${SA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"

# 1. Create the Service Account if it doesn't exist
if ! gcloud iam service-accounts describe ${SA_EMAIL} --project=${PROJECT_ID} &>/dev/null; then
  echo "Creating Service Account: ${SA_EMAIL}..."
  gcloud iam service-accounts create ${SA_NAME} \
    --display-name="Cert Manager Google CAS Issuer" \
    --project=${PROJECT_ID}
else
  echo "Service Account ${SA_EMAIL} already exists."
fi

# 2. Grant authoritative CAS Requester role
echo "Authorizing CAS Requester role..."
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/privateca.certificateRequester" --quiet > /dev/null

# 3. Establish Workload Identity linkage
echo "Linking Workload Identity..."
gcloud iam service-accounts add-iam-policy-binding ${SA_EMAIL} \
  --role="roles/iam.workloadIdentityUser" \
  --member="serviceAccount:${PROJECT_ID}.svc.id.goog[cert-manager/${SA_NAME}]" \
  --project=${PROJECT_ID} --quiet > /dev/null

echo "------------------------------------------------------------------------"
echo "SUCCESS: High-Assurance Services & mTLS Identity Synchronized."
echo "------------------------------------------------------------------------"
