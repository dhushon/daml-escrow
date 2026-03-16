#!/bin/bash

# Stablecoin Escrow API Orchestration Script
# Demonstrates a full "Golden Path" transaction: Create -> Get -> Release

API_URL="http://localhost:8080"

echo "--- 1. Creating a new Escrow ---"
CREATE_RES=$(curl -s -X POST "$API_URL/escrows" \
  -H "Content-Type: application/json" \
  -d '{
    "buyer": "buyer-alice",
    "seller": "seller-bob",
    "amount": 500.0,
    "currency": "USD"
  }')

echo "Response: $CREATE_RES"
ESCROW_ID=$(echo $CREATE_RES | jq -r '.id')

if [ "$ESCROW_ID" == "null" ] || [ -z "$ESCROW_ID" ]; then
  echo "Error: Failed to create escrow."
  exit 1
fi

echo -e "\n--- 2. Fetching Escrow Details ($ESCROW_ID) ---"
GET_RES=$(curl -s -X GET "$API_URL/escrows/$ESCROW_ID")
echo "Response: $GET_RES"

STATE=$(echo $GET_RES | jq -r '.state')
if [ "$STATE" != "Created" ]; then
  echo "Error: Initial state should be 'Created'."
  exit 1
fi

echo -e "\n--- 3. Releasing Escrow Funds ---"
RELEASE_RES=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/escrows/$ESCROW_ID/release")
echo "HTTP Status Code: $RELEASE_RES"

if [ "$RELEASE_RES" != "200" ]; then
  echo "Error: Release failed."
  exit 1
fi

echo -e "\n--- 4. Verifying Final State ---"
FINAL_RES=$(curl -s -X GET "$API_URL/escrows/$ESCROW_ID")
echo "Final Response: $FINAL_RES"

FINAL_STATE=$(echo $FINAL_RES | jq -r '.state')
if [ "$FINAL_STATE" != "Released" ]; then
  echo "Error: Final state should be 'Released'."
  exit 1
fi

echo -e "\n--- Golden Path Orchestration Successful! ---"
