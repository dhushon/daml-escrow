# Phase 13: Cloud-Native Persistence Layer
# Deploys a Cloud SQL for PostgreSQL instance to replace local containers.

resource "google_project_service" "sqladmin" {
  service            = "sqladmin.googleapis.com"
  disable_on_destroy = false
}

# --- Cloud SQL Instance (Zonal/Shared Core for Dev) ---

resource "google_sql_database_instance" "escrow_db" {
  name             = "escrow-db-${var.environment}"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    # db-f1-micro is the most cost-effective for dev
    tier = "db-f1-micro"
    
    # Cost-Saving: No high availability for dev
    availability_type = "ZONAL"

    backup_configuration {
      enabled = true
    }

    ip_configuration {
      ipv4_enabled    = true # Allow public IP for dev/local access
      ssl_mode        = "ENCRYPTED_ONLY"
      
      # Authoritative Dev Access: Allow the developer's current IP if needed
      # In production, this would be restricted to GKE private IPs via Private Service Connect
      authorized_networks {
        name  = "local-dev"
        value = "0.0.0.0/0" # Warning: Only for dev speed, restrict in prod
      }
    }

    location_preference {
      zone = "${var.region}-a"
    }

    database_flags {
      name  = "log_connections"
      value = "on"
    }
  }

  # Allow for quick dev iteration
  deletion_protection = false
  depends_on          = [google_project_service.sqladmin]
}

# --- User Configuration Database ---

resource "google_sql_database" "user_config" {
  name     = "user_config"
  instance = google_sql_database_instance.escrow_db.name
}

# --- DB Credentials & Secret Vending ---

resource "random_password" "db_password" {
  length  = 16
  special = true
}

resource "google_sql_user" "escrow_admin" {
  name     = "escrow_admin"
  instance = google_sql_database_instance.escrow_db.name
  password = random_password.db_password.result
}

resource "google_secret_manager_secret" "db_dsn" {
  secret_id = "user-config-dsn-${var.environment}"
  labels    = merge(var.common_labels, { svc = "escrow-api" })
  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }
}

resource "google_secret_manager_secret_version" "db_dsn_v1" {
  secret = google_secret_manager_secret.db_dsn.id
  secret_data = "postgres://escrow_admin:${random_password.db_password.result}@${google_sql_database_instance.escrow_db.public_ip_address}:5432/user_config?sslmode=verify-ca"
}

# --- IAM: Secret Access for Developer ---

resource "google_secret_manager_secret_iam_member" "db_dsn_accessor" {
  project   = var.project_id
  secret_id = google_secret_manager_secret.db_dsn.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "user:dan@vdatacloudai.com"
}
