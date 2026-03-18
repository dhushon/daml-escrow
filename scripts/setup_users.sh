#!/bin/bash
set -e

JSON_API="http://localhost:7575"

echo "Fetching parties..."
PARTIES=$(curl -s $JSON_API/v2/parties | jq -r '.partyDetails[] | .party')

CB_ID=$(echo "$PARTIES" | grep "^CentralBank::")
BUYER_ID=$(echo "$PARTIES" | grep "^Buyer::")
SELLER_ID=$(echo "$PARTIES" | grep "^Seller::")
MEDIATOR_ID=$(echo "$PARTIES" | grep "^EscrowMediator::")

echo "Mapping:"
echo "CB: $CB_ID"
echo "Buyer: $BUYER_ID"
echo "Seller: $SELLER_ID"
echo "Mediator: $MEDIATOR_ID"

create_user() {
  local user_id=$1
  local party_id=$2
  echo "Creating user $user_id..."
  curl -X POST "$JSON_API/v2/users" \
    -H "Content-Type: application/json" \
    -d "{
      \"user\": {
        \"id\": \"$user_id\",
        \"primaryParty\": \"$party_id\",
        \"isDeactivated\": false,
        \"identityProviderId\": \"\"
      }
    }" || echo "User $user_id already exists"
}

grant_rights() {
  local user_id=$1
  local party_id=$2
  echo "Granting actAs to $user_id for $party_id..."
  curl -X POST "$JSON_API/v2/users/$user_id/rights" \
    -H "Content-Type: application/json" \
    -d "{
      \"userId\": \"$user_id\",
      \"actAs\": [\"$party_id\"],
      \"readAs\": [],
      \"identityProviderId\": \"\"
    }"
}

create_user "CentralBank" "$CB_ID"
create_user "Buyer" "$BUYER_ID"
create_user "Seller" "$SELLER_ID"
create_user "EscrowMediator" "$MEDIATOR_ID"

grant_rights "CentralBank" "$CB_ID"
grant_rights "Buyer" "$BUYER_ID"
grant_rights "Buyer" "$CB_ID" # Buyer acts as CB for payment simulation
grant_rights "Seller" "$SELLER_ID"
grant_rights "EscrowMediator" "$MEDIATOR_ID"

echo "User setup complete."
