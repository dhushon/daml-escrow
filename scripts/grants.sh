#!/bin/bash
# high-assurance contributor provisioning script
# authoritatively grants institutional roles for gcp project vdcai-daml

PROJECT_ID="vdcai-daml"
CONTRIBUTOR="dan@vdatacloudai.com"
ENVIRONMENT="dev"

echo "------------------------------------------------------------------------"
echo "PROVISIONING INSTITUTIONAL CONTRIBUTOR: $CONTRIBUTOR"
echo "PROJECT: $PROJECT_ID | ENV: $ENVIRONMENT"
echo "------------------------------------------------------------------------"

# 1. SECRET MANAGER: Authoritative Vending & Auditing
echo "[1/4] Granting Secret Manager Privileges..."
for secret in okta-client-secret bitgo-access-token circle-api-key; do
  FULL_SECRET="${secret}-${ENVIRONMENT}"
  
  # secretAccessor: Allows reading the actual payload (Versions)
  gcloud secrets add-iam-policy-binding $FULL_SECRET \
    --member="user:$CONTRIBUTOR" \
    --role="roles/secretmanager.secretAccessor" \
    --project=$PROJECT_ID --quiet > /dev/null

  # viewer: Allows viewing metadata (required for Terraform state refresh)
  gcloud secrets add-iam-policy-binding $FULL_SECRET \
    --member="user:$CONTRIBUTOR" \
    --role="roles/secretmanager.viewer" \
    --project=$PROJECT_ID --quiet > /dev/null
done

# 2. ARTIFACT REGISTRY: High-Assurance Release Engineering
echo "[2/4] Granting Artifact Registry Privileges..."
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="user:$CONTRIBUTOR" \
  --role="roles/artifactregistry.admin" --quiet > /dev/null

# 3. GKE: Kubernetes Orchestration
echo "[3/4] Granting GKE Admin Privileges..."
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="user:$CONTRIBUTOR" \
  --role="roles/container.admin" --quiet > /dev/null

# 4. COMPUTE: Network & VPC Auditing
echo "[4/4] Granting Network Auditing Privileges..."
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="user:$CONTRIBUTOR" \
  --role="roles/compute.networkAdmin" --quiet > /dev/null

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="user:$CONTRIBUTOR" \
  --role="roles/compute.publicIpAdmin" --quiet > /dev/null

echo "------------------------------------------------------------------------"
echo "SUCCESS: High-Assurance Contributor Privileges Provisioned."
echo "------------------------------------------------------------------------"
\n# 5. DOCKER: High-Assurance Release Engineering\necho "[5/5] Configuring Docker for vdcai-daml..."\ngcloud auth configure-docker us-central1-docker.pkg.dev --project=$PROJECT_ID --quiet > /dev/null
