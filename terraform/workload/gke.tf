# Phase 8.2: Cost-Optimized GKE Infrastructure
# Deploys a zonal GKE cluster using Spot instances to minimize development costs.

resource "google_project_service" "container" {
  service            = "container.googleapis.com"
  disable_on_destroy = false
}

# --- GKE Cluster (Zonal for cost savings) ---

resource "google_container_cluster" "primary" {
  name     = "escrow-cluster-${var.environment}"
  location = "${var.region}-a" # Single zone to avoid regional fees
  
  network    = google_compute_network.main.id
  subnetwork = google_compute_subnetwork.public.id

  # High-Assurance: Separate node pool management
  remove_default_node_pool = true
  initial_node_count       = 1

  # Cost-Saving: Enable deletion protection for production, but disable for dev
  deletion_protection = false

  # High-Assurance Networking
  ip_allocation_policy {
    cluster_secondary_range_name  = "" # Use default auto-generated ranges
    services_secondary_range_name = ""
  }

  # High-Assurance: Enable Workload Identity
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  depends_on = [google_project_service.container]
}

# --- Spot Node Pool (Institutional Grade / Developer Cost) ---

resource "google_container_node_pool" "spot_nodes" {
  name       = "spot-node-pool-${var.environment}"
  location   = google_container_cluster.primary.location
  cluster    = google_container_cluster.primary.name
  node_count = 1

  autoscaling {
    min_node_count = 1
    max_node_count = 3
  }

  node_config {
    # e2-standard-4 provides 4 vCPU / 16GB RAM - mandatory for Tripartite Canton workloads
    machine_type = "e2-standard-4"
    
    # Cost-Saving: Use SPOT instances (up to 90% cheaper)
    spot = true

    # IAM: Standard GKE permissions
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    labels = merge(var.common_labels, {
      env  = var.environment
      pool = "spot"
    })

    # Allow firewall rules to target these nodes
    tags = ["dev-node", "gke-node"]
    
    # Metadata for better logging
    metadata = {
      disable-legacy-endpoints = "true"
    }
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }
}

# --- Outputs for K8s Context ---

output "kubernetes_cluster_name" {
  value = google_container_cluster.primary.name
}

output "kubernetes_cluster_host" {
  value     = google_container_cluster.primary.endpoint
  sensitive = true
}
