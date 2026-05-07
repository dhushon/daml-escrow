# Phase 8.3: Institutional Networking & DNS
# Provisions the public entrypoint for the Stablecoin Escrow platform.
# High-Assurance: Conditional provisioning to minimize development costs.

# --- 1. Global Static IP (Unified Gateway) ---

resource "google_compute_global_address" "escrow_gateway_ip" {
  count = var.enable_static_ip ? 1 : 0
  name  = "escrow-gateway-ip-${var.environment}"
  
  labels = merge(var.common_labels, {
    env = var.environment
  })
}

# --- 2. Cloud DNS: Institutional Mapping ---

resource "google_dns_managed_zone" "escrow_zone" {
  count       = var.enable_static_ip ? 1 : 0
  name        = "vdatacloudai-zone"
  dns_name    = "vdatacloudai.com."
  description = "Authoritative zone for institutional escrow services"

  labels = merge(var.common_labels, {
    env = var.environment
  })
}

resource "google_dns_record_set" "api_vdatacloudai" {
  count        = var.enable_static_ip ? 1 : 0
  name         = "api.vdatacloudai.com."
  managed_zone = google_dns_managed_zone.escrow_zone[0].name
  type         = "A"
  ttl          = 300
  rrdatas      = [google_compute_global_address.escrow_gateway_ip[0].address]
}

# --- Outputs for GKE Ingress ---

output "gateway_static_ip" {
  value = var.enable_static_ip ? google_compute_global_address.escrow_gateway_ip[0].address : "dynamic"
}
