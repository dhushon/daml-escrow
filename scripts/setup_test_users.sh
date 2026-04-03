#!/bin/bash
# setup_test_users.sh - Pre-provisions consistent personas for Phase 9 testing.

echo "--- Okta Identity Provisioning: Consistent Personas ---"
echo "Please create the following users in your Okta Admin Console."
echo "CRITICAL: Ensure 'User must change password on first login' is UNCHECKED."
echo ""

PASSWORD="Stablecoin2026!"

printf "%-20s | %-30s | %-20s\n" "Name" "Email" "Group"
echo "--------------------------------------------------------------------------------"
printf "%-20s | %-30s | %-20s\n" "Joey Buyer" "joey@buyer.com" "EscrowBuyers"
printf "%-20s | %-30s | %-20s\n" "Jimmy Seller" "jimmy@seller.com" "EscrowSellers"
printf "%-20s | %-30s | %-20s\n" "Sally Mediator" "sally@mediator.com" "EscrowMediators"
printf "%-20s | %-30s | %-20s\n" "Bob Banker" "bob@banker.com" "EscrowBank"
printf "%-20s | %-30s | %-20s\n" "Invited Seller" "invited-seller@vdatacloud.com" "None (Guest)"
echo "--------------------------------------------------------------------------------"
echo "Common Password: $PASSWORD"
echo ""
echo "Identity Verification Workflow:"
echo "1. Create the users above manually or via Okta API."
echo "2. Assign them to the respective groups created by Terraform."
echo "3. Run the API with: bin/escrow-api serve --env dev --bypass=false"
echo "4. Authenticate as Joey to verify JIT provisioning onto the DAML ledger."
