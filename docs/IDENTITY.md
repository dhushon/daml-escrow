# High-Assurance Identity Registry

This document authoritatively defines the security principals and institutional roles governing the Stablecoin Escrow platform.

## 1. Contributor Principal (Developer/Admin)

The **Contributor Principal** is the primary identity authoritatively permitted to provision infrastructure and manage the high-assurance release cycle.

*   **Principal ID**: `dan@vdatacloudai.com`
*   **Status**: Authoritative
*   **Mandatory Roles**:
    *   `roles/container.admin`: Authoritative GKE orchestration.
    *   `roles/artifactregistry.admin`: High-assurance repository management.
    *   `roles/compute.admin`: Sovereign VPC and Static IP provisioning.
    *   `roles/secretmanager.viewer`: Institutional KeyVault auditing.
    *   `roles/resourcemanager.projectIamAdmin`: Cross-service identity synchronization.

## 2. GKE Deployment Service Principal

The **Deployment Service Principal** is the high-assurance machine identity utilized by the GKE cluster to pull sovereign images and access cloud-native secrets.

*   **Principal ID**: `840215557991-compute@developer.gserviceaccount.com` (Default GKE/Compute SA)
*   **Status**: Authoritative
*   **Authorized Access**:
    *   `roles/artifactregistry.reader`: Permissioned to pull production-grade images from `vdcai-daml`.
    *   `roles/secretmanager.secretAccessor`: (Planned) Permissioned to authoritatively vend Okta and Stablecoin tokens.

## 3. Platform Administrator (Global Root)

The **Platform Administrator** serves as the root of trust, authoritatively managing project-level billing, service enablement, and principal enrollment.

*   **Model**: IAM Group or Sovereign Identity.
*   **Scope**: Global `vdcai-daml` project.

## High-Assurance Governance Policy

1.  **Explicit Enrollment**: All principals MUST be authoritatively provisioned via `scripts/grants.sh` or `terraform/iam.tf`.
2.  **Least Privilege**: Access is authoritatively scoped to the `-dev` and `-prod` environments respectively.
3.  **Auditability**: Every action taken by a principal is authoritatively recorded in the SOC2-compliant **Immutable Audit Log Sink**.
