# Identity & Infrastructure Plan: Phase 6 Prep

## 1. Google Cloud Identity Platform (GCIP) - Terraform Structure
Focus: Initializing the IdP and configuring the Home Realm Discovery (HRD) foundation.

```hcl
# main.tf (Skeleton)
resource "google_identity_platform_project_default_config" "default" {
  project = var.project_id
}

# Retail OIDC (Google)
resource "google_identity_platform_default_supported_idp_config" "google" {
  enabled = true
  idp_id  = "google.com"
  client_id     = var.google_client_id
  client_secret = var.google_client_secret
}

# Enterprise SAML Tenant Template
resource "google_identity_platform_tenant" "enterprise_tenant" {
  for_each     = var.enterprise_clients
  display_name = each.value.name
  project      = var.project_id
}
```

## 2. Home Realm Discovery (HRD) Logic
**Flow:**
1. User enters email on `onboard.astro` or `login.astro`.
2. Frontend calls `GET /api/v1/auth/discovery?email=user@company.com`.
3. Backend checks `domain_mappings` table:
   - If `company.com` -> Return `SAML` + `TenantID`.
   - If `gmail.com` -> Return `OIDC` (Google).
4. Frontend redirects to the appropriate provider.

## 3. Pluggable KYC (Mock Implementation)
**Interface:**
```go
type ComplianceService interface {
    VerifyIdentity(ctx context.Context, identity UserIdentity) (bool, error)
}

// MockCompliance always returns true for Phase 6.
type MockCompliance struct{}
func (m *MockCompliance) VerifyIdentity(...) (bool, error) { return true, nil }
```

## 4. Invitation Logic: Select vs. Create
- **Existing Account:** If the authenticated `sub` already has a `DamlUserID` mapping, the invitation links the existing Party.
- **Novel Account:** If the `sub` is new, the system allocates a `partyIdHint` based on the email domain (e.g. `Buyer_DataCloud_xyz`) and provisions a new Daml User.
