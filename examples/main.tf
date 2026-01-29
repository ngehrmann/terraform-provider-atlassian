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
  api_token = var.atlassian_api_token # Must be Atlassian Admin API token
  org_id    = var.atlassian_org_id    # Organization ID from Atlassian Admin

  # Optional parameters
  base_url = "https://api.atlassian.com"
}

# Create a test team
resource "atlassian_team" "test_team" {
  display_name = "Test Team"
  description  = "A test team created via Terraform"
  team_type    = "OPEN"                # Valid values: OPEN or CLOSED
  site_id      = var.atlassian_site_id # Optional
}
