# Phase 7.1: Institutional Compliance & Auditing
# Provision GCP Certificate Authority and Immutable Audit Logging.

# --- 1. Certificate Authority Service (for internal mTLS) ---

resource "google_privateca_ca_pool" "escrow_ca_pool" {
  name     = "escrow-ca-pool-${var.environment}"
  location = var.region
  tier     = "ENTERPRISE"
  
  labels = merge(var.common_labels, {
    env = var.environment
  })
}

resource "google_privateca_certificate_authority" "root_ca" {
  location                 = var.region
  pool                     = google_privateca_ca_pool.escrow_ca_pool.name
  certificate_authority_id = "escrow-root-ca-${var.environment}"
  
  config {
    subject_config {
      subject {
        organization = "VDataCloud AI"
        common_name  = "escrow-root-ca"
      }
    }
    x509_config {
      ca_options {
        is_ca = true
      }
      key_usage {
        base_key_usage {
          cert_sign = true
          crl_sign  = true
        }
        extended_key_usage {
          server_auth = true
          client_auth = true
        }
      }
    }
  }

  key_spec {
    algorithm = "RSA_PKCS1_4096_SHA256"
  }

  # High-Assurance: We let Google manage the HSM for the root key
  deletion_protection = true
}

# --- 2. Immutable Audit Logging (SOC2 Compliance) ---

resource "google_storage_bucket" "audit_logs" {
  name          = "escrow-audit-logs-${var.project_id}-${var.environment}"
  location      = var.region
  force_destroy = false
  
  storage_class = "ARCHIVE"

  # SOC2 Requirement: Immutable storage (Object Locking)
  retention_policy {
    is_locked        = true
    retention_period = 220752000 # 7 Years (Institutional Standard)
  }

  uniform_bucket_level_access = true

  labels = merge(var.common_labels, {
    env = var.environment
  })
}

resource "google_logging_project_sink" "audit_sink" {
  name        = "escrow-compliance-sink-${var.environment}"
  description = "SOC2 immutable export of all administrative and data access logs"
  destination = "storage.googleapis.com/${google_storage_bucket.audit_logs.name}"

  # High-Assurance: Capture everything
  filter = "logName:\"logs/cloudaudit.googleapis.com\""

  unique_writer_identity = true
}

# Grant the sink writer identity permission to write to the bucket
resource "google_storage_bucket_iam_member" "sink_writer" {
  bucket = google_storage_bucket.audit_logs.name
  role   = "roles/storage.objectCreator"
  member = google_logging_project_sink.audit_sink.writer_identity
}
