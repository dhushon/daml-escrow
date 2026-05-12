#!/bin/bash
# setup_test_users.sh - Pre-provisions consistent personas for Phase 9 testing.

echo "--- Okta Identity Provisioning: Automated via Terraform ---"
echo "Applying infrastructure changes to create test users..."
echo ""

cd terraform && terraform apply -auto-approve

echo ""
echo "Identity Verification Workflow:"
echo "1. Create the users above manually or via Okta API."
echo "2. Assign them to the respective groups created by Terraform."
echo "3. Run the API with: bin/escrow-api serve --env dev --bypass=false"
echo "4. Authenticate as Joey to verify JIT provisioning onto the DAML ledger."
