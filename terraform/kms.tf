# Phase 7.2: High-Assurance Key Management
# Provision GCP KMS for hardware-backed (HSM) signing operations.

# --- Required APIs ---

resource "google_project_service" "kms" {
  service            = "cloudkms.googleapis.com"
  disable_on_destroy = false
}

# --- Key Ring ---

resource "google_kms_key_ring" "escrow_keyring" {
  name       = "escrow-keyring-${var.environment}"
  location   = var.region
  depends_on = [google_project_service.kms]
}

# --- Asymmetric Signing Key (Oracle Simulator) ---

resource "google_kms_crypto_key" "oracle_signer" {
  name     = "oracle-signer-key-${var.environment}"
  key_ring = google_kms_key_ring.escrow_keyring.id
  purpose  = "ASYMMETRIC_SIGN"

  version_template {
    algorithm        = "EC_SIGN_P256_SHA256"
    protection_level = "HSM" # Authoritative Hardware-Backed Protection
  }

  labels = merge(var.common_labels, {
    env = var.environment
    svc = "oracle-simulator"
  })

  lifecycle {
    prevent_destroy = false # Allow for dev iteration, set to true for prod
  }
}

# --- IAM: Allow the developer persona to sign for local pre-test ---

resource "google_kms_crypto_key_iam_member" "dev_signer" {
  crypto_key_id = google_kms_crypto_key.oracle_signer.id
  role          = "roles/cloudkms.signerVerifier"
  member        = "user:dan@vdatacloudai.com"
}

resource "google_kms_crypto_key_iam_member" "dev_viewer" {
  crypto_key_id = google_kms_crypto_key.oracle_signer.id
  role          = "roles/viewer"
  member        = "user:dan@vdatacloudai.com"
}
