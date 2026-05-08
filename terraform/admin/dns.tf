# Phase 8.3: Institutional Networking & DNS
# Provisions the public entrypoint for the Stablecoin Escrow platform.
# High-Assurance: Regional Static IP for GKE Nginx Ingress.

# --- 1. Regional Static IP (Nginx LoadBalancer) ---

resource "google_compute_address" "escrow_gateway_ip" {
  count  = var.disable_static_ip ? 0 : 1
  name   = "escrow-gateway-ip-${var.environment}"
  region = var.region
  
  labels = merge(var.common_labels, {
    env = var.environment
  })
}

# --- 2. Cloud DNS: Institutional Mapping ---

resource "google_dns_managed_zone" "escrow_zone" {
  count       = var.disable_static_ip ? 0 : 1
  name        = "vdatacloudai-zone"
  dns_name    = "vdatacloudai.com."
  description = "Authoritative zone for institutional escrow services"

  labels = merge(var.common_labels, {
    env = var.environment
  })
}

resource "google_dns_record_set" "api_vdatacloudai" {
  count        = var.disable_static_ip ? 0 : 1
  name         = "api.vdatacloudai.com."
  managed_zone = google_dns_managed_zone.escrow_zone[0].name
  type         = "A"
  ttl          = 300
  rrdatas      = [google_compute_address.escrow_gateway_ip[0].address]
}

# --- Outputs for GKE Ingress ---

output "gateway_static_ip" {
  value = var.disable_static_ip ? "dynamic" : google_compute_address.escrow_gateway_ip[0].address
}
