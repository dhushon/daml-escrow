# Google Cloud Identity Platform (GCIP) Setup Guide

This guide details the manual and automated steps required to configure the high-assurance identity infrastructure for the Stablecoin Escrow platform.

## 1. Prerequisites & GCP Preparation

### Enable the API
Before Terraform can manage your identity resources, you must manually enable the service in the GCP Console:
1.  Go to the **[Identity Platform Marketplace Page](https://console.cloud.google.com/market/details/google-cloud-api/identitytoolkit.googleapis.com)**.
2.  Select your project and click **Enable**.
3.  This initializes the default project configuration required for multi-tenancy.

### Authenticate Terraform
Ensure your local environment has permission to provision project-level resources:
```bash
gcloud auth application-default login
```

---

## 2. Creating OAuth 2.0 Credentials (OIDC)

To enable "Sign in with Google," you must create a Client ID that the platform will use to verify users.

1.  Navigate to **[APIs & Services > Credentials](https://console.cloud.google.com/apis/credentials)**.
2.  Click **+ Create Credentials** > **OAuth client ID**.
3.  **Application Type:** Select `Web application`.
4.  **Name:** `Stablecoin Escrow Platform`.
5.  **Authorized JavaScript origins:**
    *   `http://localhost:4321` (Default Astro port)
    *   `http://localhost:8080` (Standard fallback)
6.  **Authorized redirect URIs:**
    *   `http://localhost:4321`
    *   `http://localhost:8080`
    *   `https://<YOUR_PROJECT_ID>.firebaseapp.com/__/auth/handler` (Crucial: Required by GCIP)
7.  **Save:** Note your `Client ID` and `Client Secret`. These go into your `terraform.tfvars`.

---

## 3. Local Development Whitelisting

Identity Platform enforces strict origin checks. You must whitelist `localhost` to allow the login SDK to function during development.

1.  Go to **[Identity Platform > Settings](https://console.cloud.google.com/customer-identity/settings)**.
2.  Select the **Security** tab.
3.  Under **Authorized Domains**, click **Add Domain**.
4.  Add `localhost`.
5.  Click **Save**.

---

## 4. Infrastructure Deployment (Terraform)

Once the manual credentials and whitelisting are complete, use Terraform to provision the Enterprise Tenants and SAML providers.

### Configuration
1.  Enter the terraform directory: `cd terraform`
2.  Create your variables file: `cp terraform.tfvars.example terraform.tfvars`
3.  Edit `terraform.tfvars` with the IDs obtained in Step 2.

### Execution
```bash
terraform init
terraform apply
```

### Capturing Output
After a successful apply, Terraform will output a `tenant_ids` block. **Save these IDs**; they must be added to `config/identity_providers.yaml` to enable Home Realm Discovery (HRD).

---

## 5. Summary Checklist for .env
Ensure your root `.env` file matches your GCP settings:
```bash
# Obtain from Step 2
AUTH_CLIENT_ID="your-id.apps.googleusercontent.com"

# Project ID from GCP Console
AUTH_AUDIENCE="https://<YOUR_PROJECT_ID>.firebaseapp.com"
AUTH_ISSUER="https://securetoken.google.com/<YOUR_PROJECT_ID>"
```
