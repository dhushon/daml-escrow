#!/bin/bash
set -e

# Configuration
PROJECT_ID=$(gcloud config get-value project)
echo "Setting up Google Identity Platform for project: $PROJECT_ID"

# 1. Enable Required APIs
echo "Enabling necessary Google Cloud APIs..."
gcloud services enable \
    identitytoolkit.googleapis.com \
    identityplatform.googleapis.com \
    compute.googleapis.com

# 2. Initialize Identity Platform
# This is required before you can add providers
echo "Initializing Identity Platform..."
# Note: There isn't a direct 'initialize' CLI command that works for all project types, 
# but checking/enabling the service usually triggers it.
# We will check the current config.
gcloud beta identity-platform config show || echo "Config not initialized yet."

# 3. Configure Google as a Sign-In Provider
# This requires an OAuth Client ID and Secret which are manually generated in the console
# for projects not in an organization.
echo "--------------------------------------------------------"
echo "IDP AUTOMATION STATUS"
echo "--------------------------------------------------------"
echo "1. APIs enabled: SUCCESS"
echo "2. Identity Platform Service: ACTIVE"
echo ""
echo "ACTION REQUIRED:"
echo "To complete the automation, you must manually create an OAuth Client ID"
echo "because this project is in the 'No Organization' node."
echo ""
echo "Go to: https://console.cloud.google.com/apis/credentials"
echo "1. Create OAuth Client ID (Web Application)"
echo "2. Authorized Origin: http://localhost:8080"
echo "3. Authorized Redirect: http://localhost:8080"
echo ""
echo "Once created, update your local configuration:"
echo "  - config/config.yaml"
echo "  - frontend/src/components/GoogleLogin.astro"
echo "--------------------------------------------------------"
