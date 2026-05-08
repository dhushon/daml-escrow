# Phase 9: High-Assurance Identity (Okta Integration)
# authoritatively manages institutional principals and OIDC federation.

# --- 1. OIDC Application: The Institutional Portal ---

# resource "okta_app_oauth" "escrow_portal" {
  count          = var.enable_okta_idp ? 1 : 0
  label          = "Stablecoin Escrow Portal (${var.environment})"
  type           = "browser"
  grant_types    = ["authorization_code", "refresh_token"]
  redirect_uris  = ["http://localhost:4321/auth/callback", "https://api.vdatacloudai.com/auth/callback"]
  response_types = ["code"]
  
  issuer_mode   = "ORG_URL"
  token_endpoint_auth_method = "none" # PKCE for browser-based portal
}

# --- 2. Authorization Server: Scoped Institutional Authority ---

# resource "okta_auth_server" "escrow_server" {
  count       = var.enable_okta_idp ? 1 : 0
  name        = "Stablecoin Escrow Authority (${var.environment})"
  audiences   = ["daml-escrow"]
  description = "Authoritative server for institutional escrow scopes"
}

# resource "okta_auth_server_scope" "escrow_scopes" {
  for_each       = var.enable_okta_idp ? toset(["escrow:read", "escrow:write", "escrow:accept", "system:admin"]) : []
  auth_server_id = okta_auth_server.escrow_server[0].id
  name           = each.key
  description    = "Authoritative institutional scope: ${each.key}"
}
