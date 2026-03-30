# 1. Initialize Google Identity Platform
resource "google_identity_platform_config" "default" {
  provider = google-beta
  project  = var.project_id
}

# 2. Configure Retail OIDC (Google)
resource "google_identity_platform_default_supported_idp_config" "google" {
  provider      = google-beta
  project       = var.project_id
  enabled       = true
  idp_id        = "google.com"
  client_id     = var.google_client_id
  client_secret = var.google_client_secret
}

# 3. Provision Enterprise SAML Tenants
resource "google_identity_platform_tenant" "enterprise_tenant" {
  provider     = google-beta
  for_each     = var.enterprise_clients
  project      = var.project_id
  display_name = each.value.name
  
  allow_password_signup = true
}

# 4. Example SAML IdP Config for DataCloud
resource "google_identity_platform_tenant_saml_idp_config" "datacloud_saml" {
  provider     = google-beta
  project      = var.project_id
  name         = "saml.datacloud" # Must start with 'saml.'
  tenant       = google_identity_platform_tenant.enterprise_tenant["datacloud"].name
  display_name = "DataCloud SSO"
  enabled      = true
  idp_id       = "saml.datacloud" # Logical ID for HRD
  
  idp_entity_id = "https://sts.windows.net/datacloud-uuid/"
  sso_url       = "https://login.microsoftonline.com/datacloud-uuid/saml2"
  
  idp_certificates {
    x509_certificate = "MII..." # Placeholder
  }

  sp_config {
    sp_entity_id = "https://escrow-platform.firebaseapp.com"
    callback_uri = "https://escrow-platform.firebaseapp.com/__/auth/handler"
  }
}

# 5. Output Tenant IDs for Go API Consumption
output "tenant_ids" {
  value = {
    for k, v in google_identity_platform_tenant.enterprise_tenant : k => v.name
  }
}
