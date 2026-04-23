# Phase 13: Development-Ready Networking
# Establishes a VPC with "Development Rules" allowing for local internet access
# while providing the foundation for future institutional GKE nodes.

resource "google_compute_network" "main" {
  name                    = "escrow-vpc-${var.environment}"
  auto_create_subnetworks = false
  routing_mode            = "REGIONAL"
}

resource "google_compute_subnetwork" "public" {
  name          = "escrow-public-${var.environment}"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.main.id

  # Enables future private connectivity to Google APIs
  private_ip_google_access = true
}

# --- Development Firewall Rules ---

resource "google_compute_firewall" "allow_local_dev" {
  name    = "allow-local-dev-${var.environment}"
  network = google_compute_network.main.name

  # Allow standard web and API traffic for local pre-testing
  allow {
    protocol = "tcp"
    ports    = ["80", "443", "8080", "8081", "7575"]
  }

  source_ranges = ["0.0.0.0/0"] # Open for development as requested
  target_tags   = ["dev-node"]
}

resource "google_compute_firewall" "allow_internal" {
  name    = "allow-internal-${var.environment}"
  network = google_compute_network.main.name

  allow {
    protocol = "icmp"
  }

  allow {
    protocol = "tcp"
    ports    = ["0-65535"]
  }

  source_ranges = ["10.0.0.0/24"]
}
