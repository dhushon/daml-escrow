variable "project_id" {
  description = "The GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "environment" {
  description = "Deployment environment (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "common_labels" {
  description = "Common labels for all resources"
  type        = map(string)
  default     = {
    managed-by = "terraform"
    project    = "daml-escrow"
  }
}
