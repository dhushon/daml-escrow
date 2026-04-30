# Phase 12: High-Assurance Secret Management
# Provision GCP Secret Manager for critical institutional credentials.

# --- Required APIs ---

resource "google_project_service" "secretmanager" {
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

# --- Secrets ---

resource "google_secret_manager_secret" "okta_client_secret" {
  secret_id  = "okta-client-secret-${var.environment}"
  depends_on = [google_project_service.secretmanager]

  labels = merge(var.common_labels, {
    env = var.environment
    svc = "escrow-api"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "bitgo_access_token" {
  secret_id  = "bitgo-access-token-${var.environment}"
  depends_on = [google_project_service.secretmanager]

  labels = merge(var.common_labels, {
    env = var.environment
    svc = "escrow-api"
  })

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret" "circle_api_key" {
  secret_id  = "circle-api-key-${var.environment}"
  depends_on = [google_project_service.secretmanager]

  labels = merge(var.common_labels, {
    env = var.environment
    svc = "escrow-api"
  })

  replication {
    auto {}
  }
}

# --- IAM: Allow the developer persona to manage/read for local pre-test ---
# Note: In production, this would be restricted to a specific GKE Service Account.

resource "google_secret_manager_secret_iam_member" "dev_accessor" {
  for_each = toset([
    google_secret_manager_secret.okta_client_secret.secret_id,
    google_secret_manager_secret.bitgo_access_token.secret_id,
    google_secret_manager_secret.circle_api_key.secret_id
  ])

  secret_id = each.key
  role      = "roles/secretmanager.secretAccessor"
  member    = "user:${var.project_id}@gmail.com" # Placeholder: Usually current authorized user
}
