variable "project_id" {
  description = "Google Cloud Project ID"
  type        = string
}

variable "region" {
  description = "Google Cloud Region"
  type        = string
  default     = "us-central1"
}

variable "google_client_id" {
  description = "Google OIDC Client ID"
  type        = string
}

variable "google_client_secret" {
  description = "Google OIDC Client Secret"
  type        = string
  sensitive   = true
}

variable "enterprise_clients" {
  description = "Map of enterprise clients for SAML tenants"
  type = map(object({
    name   = string
    domain = string
  }))
  default = {
    "datacloud" = {
      name   = "DataCloud LLC"
      domain = "datacloud.com"
    },
    "bank" = {
      name   = "Global Bank Corp"
      domain = "bank.com"
    }
  }
}

# --- Okta Variables ---

variable "okta_org_name" {
  description = "Okta Organization Name (e.g., dev-123456)"
  type        = string
}

variable "okta_base_url" {
  description = "Okta Base URL (e.g., okta.com, oktapreview.com)"
  type        = string
  default     = "okta.com"
}

variable "okta_api_token" {
  description = "Okta API Token"
  type        = string
  sensitive   = true
}

# --- Feature Toggles ---

variable "enable_google_idp" {
  description = "Toggle to enable/disable Google Cloud Identity Platform resources"
  type        = bool
  default     = true
}

variable "enable_okta_idp" {
  description = "Toggle to enable/disable Okta Identity Provider resources"
  type        = bool
  default     = true
}
