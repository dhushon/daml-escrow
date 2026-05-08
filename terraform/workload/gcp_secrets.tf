# Phase 9.1: GCP Secret Manager (Workload Secrets)
# authoritatively manages external API keys and OIDC client secrets.

resource "google_secret_manager_secret" "okta_client_secret" {
  secret_id = "okta-client-secret-${var.environment}"
  labels    = merge(var.common_labels, { svc = "escrow-api" })
  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

resource "google_secret_manager_secret" "bitgo_access_token" {
  secret_id = "bitgo-access-token-${var.environment}"
  labels    = merge(var.common_labels, { svc = "escrow-api" })
  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

resource "google_secret_manager_secret" "circle_api_key" {
  secret_id = "circle-api-key-${var.environment}"
  labels    = merge(var.common_labels, { svc = "escrow-api" })
  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

# --- High-Assurance IAM: Secret Access for Principal ---

locals {
  workload_secrets = {
    okta   = google_secret_manager_secret.okta_client_secret.secret_id
    bitgo  = google_secret_manager_secret.bitgo_access_token.secret_id
    circle = google_secret_manager_secret.circle_api_key.secret_id
  }
}

resource "google_secret_manager_secret_iam_member" "dev_accessor" {
  for_each  = local.workload_secrets
  project   = var.project_id
  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "user:dan@vdatacloudai.com"
}
