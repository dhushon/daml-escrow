# 1. Initialize Google Identity Platform
resource "google_identity_platform_config" "default" {
  count    = var.enable_google_idp ? 1 : 0
  provider = google-beta
  project  = var.project_id
}

# 2. Configure Retail OIDC (Google)
resource "google_identity_platform_default_supported_idp_config" "google" {
  count         = var.enable_google_idp ? 1 : 0
  provider      = google-beta
  project       = var.project_id
  enabled       = true
  idp_id        = "google.com"
  client_id     = var.google_client_id
  client_secret = var.google_client_secret
}

# 3. Provision Enterprise Tenants
resource "google_identity_platform_tenant" "enterprise_tenant" {
  for_each     = var.enable_google_idp ? var.enterprise_clients : {}
  provider     = google-beta
  project       = var.project_id
  display_name = each.value.name
  
  allow_password_signup = true
}

# 5. Output Tenant IDs for Go API Consumption
output "tenant_ids" {
  value = var.enable_google_idp ? {
    for k, v in google_identity_platform_tenant.enterprise_tenant : k => v.name
  } : {}
}
