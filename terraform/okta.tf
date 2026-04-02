# --- Okta Groups & Roles ---

resource "okta_group" "buyers" {
  count       = var.enable_okta_idp ? 1 : 0
  name        = "EscrowBuyers"
  description = "Members allowed to buy assets in escrow"
}

resource "okta_group" "sellers" {
  count       = var.enable_okta_idp ? 1 : 0
  name        = "EscrowSellers"
  description = "Members allowed to sell assets in escrow"
}

resource "okta_group" "mediators" {
  count       = var.enable_okta_idp ? 1 : 0
  name        = "EscrowMediators"
  description = "Trusted third-parties allowed to resolve disputes"
}

resource "okta_group" "bank" {
  count       = var.enable_okta_idp ? 1 : 0
  name        = "EscrowBank"
  description = "Institutional members (Issuer) for settlement and disbursement"
}

# --- OIDC Application (Single Page App / Web) ---

resource "okta_app_oauth" "escrow_platform" {
  count                      = var.enable_okta_idp ? 1 : 0
  label                      = "Stablecoin Escrow Platform"
  type                       = "browser" # Represents SPA
  grant_types                = ["authorization_code", "refresh_token"]
  response_types             = ["code"]
  token_endpoint_auth_method = "none" # Public client (SPA)
  
  redirect_uris = [
    "http://localhost:4321",         # Astro Dev Port
    "http://localhost:4321/callback", 
    "http://localhost:8080"          # Fallback
  ]

  post_logout_redirect_uris = [
    "http://localhost:4321"
  ]
}

# --- Custom Authorization Server ---

resource "okta_auth_server" "escrow_auth" {
  count       = var.enable_okta_idp ? 1 : 0
  name        = "Stablecoin Escrow Auth Server"
  description = "Dedicated auth server for the high-assurance escrow platform"
  audiences   = ["daml-escrow"] # Matches Config.yaml audience
  issuer_mode = "ORG_URL"
}

# --- Scopes ---

resource "okta_auth_server_scope" "escrow_read" {
  count            = var.enable_okta_idp ? 1 : 0
  auth_server_id   = okta_auth_server.escrow_auth[0].id
  name             = "escrow:read"
  display_name     = "Read Escrows"
  description      = "Allows viewing escrow contracts and history"
  consent          = "IMPLICIT"
}

resource "okta_auth_server_scope" "escrow_write" {
  count            = var.enable_okta_idp ? 1 : 0
  auth_server_id   = okta_auth_server.escrow_auth[0].id
  name             = "escrow:write"
  display_name     = "Write Escrows"
  description      = "Allows proposing and creating new escrows"
  consent          = "IMPLICIT"
}

resource "okta_auth_server_scope" "escrow_accept" {
  count            = var.enable_okta_idp ? 1 : 0
  auth_server_id   = okta_auth_server.escrow_auth[0].id
  name             = "escrow:accept"
  display_name     = "Accept Escrows"
  description      = "Allows accepting proposals and releasing funds"
  consent          = "IMPLICIT"
}

resource "okta_auth_server_scope" "system_admin" {
  count            = var.enable_okta_idp ? 1 : 0
  auth_server_id   = okta_auth_server.escrow_auth[0].id
  name             = "system:admin"
  display_name     = "System Admin"
  description      = "Allows administrative actions and settlement finalized"
  consent          = "IMPLICIT"
}

# --- Claims (Identifying Origin) ---

resource "okta_auth_server_claim" "origin_domain" {
  count          = var.enable_okta_idp ? 1 : 0
  auth_server_id = okta_auth_server.escrow_auth[0].id
  name           = "origin_domain"
  value          = "String.substringAfter(user.email, \"@\")"
  value_type     = "EXPRESSION"
  claim_type     = "RESOURCE" # Included in the Access Token
}

# --- Authorization Policies & Rules ---

resource "okta_auth_server_policy" "default_policy" {
  count            = var.enable_okta_idp ? 1 : 0
  auth_server_id   = okta_auth_server.escrow_auth[0].id
  name             = "Default Escrow Policy"
  description      = "Allows OIDC app to request escrow scopes"
  priority         = 1
  client_whitelist = [okta_app_oauth.escrow_platform[0].id]
}

resource "okta_auth_server_policy_rule" "buyer_rule" {
  count                = var.enable_okta_idp ? 1 : 0
  auth_server_id       = okta_auth_server.escrow_auth[0].id
  policy_id            = okta_auth_server_policy.default_policy[0].id
  name                 = "Buyer Access"
  priority             = 1
  group_whitelist      = [okta_group.buyers[0].id]
  grant_type_whitelist = ["authorization_code"]
  scope_whitelist      = ["openid", "profile", "email", "escrow:read", "escrow:write", "escrow:accept"]
}

resource "okta_auth_server_policy_rule" "seller_rule" {
  count                = var.enable_okta_idp ? 1 : 0
  auth_server_id       = okta_auth_server.escrow_auth[0].id
  policy_id            = okta_auth_server_policy.default_policy[0].id
  name                 = "Seller Access"
  priority             = 2
  group_whitelist      = [okta_group.sellers[0].id]
  grant_type_whitelist = ["authorization_code"]
  scope_whitelist      = ["openid", "profile", "email", "escrow:read", "escrow:write", "escrow:accept"]
}

resource "okta_auth_server_policy_rule" "admin_rule" {
  count                = var.enable_okta_idp ? 1 : 0
  auth_server_id       = okta_auth_server.escrow_auth[0].id
  policy_id            = okta_auth_server_policy.default_policy[0].id
  name                 = "Admin Access"
  priority             = 3
  group_whitelist      = [okta_group.bank[0].id]
  grant_type_whitelist = ["authorization_code"]
  scope_whitelist      = ["openid", "profile", "email", "escrow:read", "escrow:write", "escrow:accept", "system:admin"]
}

# --- Outputs ---

output "okta_issuer_url" {
  value = var.enable_okta_idp ? okta_auth_server.escrow_auth[0].issuer : null
}

output "okta_client_id" {
  value = var.enable_okta_idp ? okta_app_oauth.escrow_platform[0].client_id : null
}
