# Phase 8: Artifact Registry
# Provision a high-assurance Docker repository for participant-sovereign images.

# --- Required APIs ---

resource "google_project_service" "artifactregistry" {
  service            = "artifactregistry.googleapis.com"
  disable_on_destroy = false
}

# --- Repository ---

resource "google_artifact_registry_repository" "escrow_repo" {
  location      = var.region
  repository_id = "escrow-platform-${var.environment}"
  description   = "Docker repository for high-assurance stablecoin escrow services"
  format        = "DOCKER"
  depends_on    = [google_project_service.artifactregistry]

  labels = merge(var.common_labels, {
    env = var.environment
  })
}

# --- IAM: Allow the developer to push/pull for local release engineering ---

resource "google_artifact_registry_repository_iam_member" "dev_repo_admin" {
  location   = google_artifact_registry_repository.escrow_repo.location
  repository = google_artifact_registry_repository.escrow_repo.name
  role       = "roles/artifactregistry.repoAdmin"
  member     = "user:dan@vdatacloudai.com"
}

output "registry_url" {
  value = "${google_artifact_registry_repository.escrow_repo.location}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.escrow_repo.repository_id}"
}
