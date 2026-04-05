# Developer Guide: Bypassing JWT Authentication

This document describes how to bypass OIDC/JWT authentication for local development and testing of the Stablecoin Escrow Platform.

## ⚠️ Security Warning

**NEVER** enable authentication bypass in a production environment. This feature is strictly for local development and integration testing where a live Identity Provider (IdP) is not available or desired.

---

## Configuration

To enable the authentication bypass, you must set two environment variables:

1.  `ENVIRONMENT=dev`: This signals that the application is running in a development context.
2.  `AUTH_BYPASS=true`: This explicitly enables the bypass logic within the `AuthMiddleware`.

If `ENVIRONMENT` is not set to `dev`, the `AUTH_BYPASS` flag will be ignored, and full OIDC validation will be enforced.

### Example: Running the API with Bypass

```bash
export ENVIRONMENT=dev
export AUTH_BYPASS=true
go run cmd/escrow-api/main.go
```

---

## How it Works

When the bypass is active:

1.  **OIDC Verifier Skip:** The application will skip the initialization of the OIDC provider and verifier. This allows the API to start even if your Okta instance is unreachable.
2.  **Middleware Bypass:** The `AuthMiddleware` will not look for an `Authorization: Bearer <token>` header or attempt to verify a JWT.
3.  **Identity Injection:** You can inject a specific user identity by providing the `X-Dev-User` header in your requests.

### Using the `X-Dev-User` Header

If the bypass is enabled, you can simulate different users by setting the `X-Dev-User` header to the desired username (e.g., `Buyer`, `Seller`, `CentralBank`).

*   If `X-Dev-User` is provided: The middleware will inject that user with a full set of scopes (`escrow:read`, `escrow:write`, `escrow:accept`, `system:admin`).
*   If `X-Dev-User` is NOT provided: The middleware defaults to injecting a user named `Buyer`.

### Example Request (cURL)

```bash
curl -X GET http://localhost:8081/api/v1/escrows \
     -H "X-Dev-User: Seller"
```

---

## Public Endpoints

The following endpoints are always exempt from JWT authentication, regardless of the bypass settings:

*   `/api/v1/health`: System health check.
*   `/api/v1/auth/discover`: Home Realm Discovery (HRD) endpoint.
*   `/swagger/*`: API documentation.
*   `/api/v1/invites/token/*`: Anonymous invitation claiming.

---

## Validation via Tests

The bypass logic is covered by unit tests in `internal/api/middleware_test.go`. You can run these tests to verify the behavior:

```bash
go test -v ./internal/api/ -run TestAuthMiddleware_Bypass
```
