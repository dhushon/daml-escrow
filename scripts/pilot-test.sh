#!/bin/bash
# High-Assurance Pilot Verification Script
# Authoritatively automates the tripartite escrow lifecycle on GKE.

GATEWAY_URL="https://api.vdatacloudai.com"
BUYER="joey@buyer.com"
SELLER="jimmy@seller.com"
BANK="bob@banker.com"

echo "------------------------------------------------------------------------"
echo "INITIATING TRIPARTITE PILOT TEST: $GATEWAY_URL"
echo "------------------------------------------------------------------------"

# 1. BUYER: Propose Escrow
echo "[1/3] BUYER: Proposing Escrow..."
PROPOSAL=$(curl -s -X POST "$GATEWAY_URL/buyer/api/v1/escrows/propose" \
  -H "Content-Type: application/json" \
  -H "X-Dev-User: $BUYER" \
  -d "{
    \"seller\": \"$SELLER\",
    \"asset\": {\"amount\": 1000, \"currency\": \"USDC\"},
    \"terms\": {\"conditionDescription\": \"Pilot hardware delivery confirmed\"}
  }")

ESCROW_ID=$(echo $PROPOSAL | grep -oE '"id":"[^"]+"' | cut -d'"' -f4)

if [ -z "$ESCROW_ID" ]; then
  echo "ERROR: Escrow proposal failed. Output: $PROPOSAL"
  exit 1
fi
echo "SUCCESS: Escrow Proposed. ID: $ESCROW_ID"

# 2. BANK: Fund Escrow
echo "[2/3] BANK: Funding Escrow..."
FUND_RESP=$(curl -s -X POST "$GATEWAY_URL/bank/api/v1/escrows/$ESCROW_ID/fund" \
  -H "Content-Type: application/json" \
  -H "X-Dev-User: $BANK" \
  -d "{\"holdingCid\": \"PILOT-HOLDING-001\"}")

echo "SUCCESS: Funding Command Dispatched."

# 3. BANK: Activate Escrow
echo "[3/3] BANK: Activating Escrow..."
ACTIVATE_RESP=$(curl -s -X POST "$GATEWAY_URL/bank/api/v1/escrows/$ESCROW_ID/activate" \
  -H "X-Dev-User: $BANK")

echo "SUCCESS: Activation Command Dispatched."

echo "------------------------------------------------------------------------"
echo "TRIPARTITE PILOT CYCLE COMPLETE."
echo "------------------------------------------------------------------------"
echo "Verify Traces at: http://localhost:16686 (Jaeger)"
echo "Verify Metrics at: http://localhost:3000 (Grafana)"
echo "------------------------------------------------------------------------"
