#!/bin/bash
# scripts/bootstrap_kind_dars.sh --- Bootstraps local Kind Canton nodes with DARs
set -e

INTERFACE_DAR="contracts/stablecoin-escrow-interfaces/.daml/dist/stablecoin-escrow-interfaces-1.0.0.dar"
ESCROW_DAR="contracts/stablecoin-escrow/.daml/dist/stablecoin-escrow-0.0.3.dar"
TEST_DAR="contracts/stablecoin-escrow-tests/.daml/dist/stablecoin-escrow-tests-1.0.0.dar"

function upload_dars_to_pod() {
  local pod=$1
  local ns=$2
  echo "==> Uploading DARs to $pod in namespace $ns..."
  
  # Start port-forwarding in background
  kubectl port-forward pod/$pod -n $ns 5002:5002 > /dev/null 2>&1 &
  local pf_pid=$!
  
  # Wait for port-forwarding to be active
  local attempts=0
  until nc -z localhost 5002 >/dev/null 2>&1; do
    attempts=$((attempts + 1))
    if [ $attempts -gt 15 ]; then
      echo "Error: Failed to port-forward to $pod"
      kill $pf_pid || true
      return 1
    fi
    sleep 1
  done
  
  echo "Uploading interface DAR..."
  daml ledger upload-dar --host localhost --port 5002 "$INTERFACE_DAR" --no-legacy-assistant-warning
  
  echo "Uploading implementation DAR..."
  daml ledger upload-dar --host localhost --port 5002 "$ESCROW_DAR" --no-legacy-assistant-warning
  
  echo "Uploading test DAR..."
  daml ledger upload-dar --host localhost --port 5002 "$TEST_DAR" --no-legacy-assistant-warning
  
  echo "Done uploading to $pod. Stopping port-forward (PID $pf_pid)..."
  kill $pf_pid || true
  sleep 1
}

upload_dars_to_pod "bank-ledger-0" "bank"
upload_dars_to_pod "depositor-ledger-0" "depositor"
upload_dars_to_pod "beneficiary-ledger-0" "beneficiary"

echo "All ledger nodes successfully provisioned with DAR files!"
