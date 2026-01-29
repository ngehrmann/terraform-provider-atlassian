terraform {
  required_version = ">= 1.0"
  required_providers {
    atlassian = {
      source  = "hashicorp/atlassian"
      version = "~> 0.1.0"
    }
  }
}

provider "atlassian" {
  # Required for Teams API - use Atlassian Admin API token
  api_token    = var.atlassian_api_token
  email        = var.atlassian_email
  organization = var.atlassian_organization
  org_id       = var.atlassian_org_id
  site_id      = var.atlassian_site_id
  base_url     = var.atlassian_base_url
}

# Create a test team
resource "atlassian_team" "test_team" {
  display_name = "Test Team"
  description  = "A test team created via Terraform"
  team_type    = "OPEN"                # Valid values: OPEN, MEMBER_INVITE, EXTERNAL, ORG_ADMIN_MANAGED
  site_id      = var.atlassian_site_id # Optional
}
