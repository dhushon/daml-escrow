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
