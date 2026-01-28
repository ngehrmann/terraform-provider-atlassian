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
  # Configuration can be provided via environment variables:
  # ATLASSIAN_API_TOKEN
  # ATLASSIAN_EMAIL
  # ATLASSIAN_ORGANIZATION
  # ATLASSIAN_BASE_URL (optional)

  # Or explicitly:
  # api_token    = var.atlassian_api_token
  # email        = var.atlassian_email
  # organization = var.atlassian_organization
  # base_url     = "https://api.atlassian.com"
}

# Create a development team
resource "atlassian_team" "dev_team" {
  name        = "Development Team"
  description = "Our main development team responsible for product development"
  type        = "development"

  members = [
    {
      account_id = "557058:12345678-1234-1234-1234-123456789012"
      role       = "admin"
    },
    {
      account_id = "557058:87654321-4321-4321-4321-210987654321"
      role       = "member"
    },
    {
      account_id = "557058:11111111-2222-3333-4444-555555555555"
      role       = "member"
    }
  ]
}

# Create a support team
resource "atlassian_team" "support_team" {
  name        = "Customer Support"
  description = "Team handling customer support and issues"
  type        = "support"

  members = [
    {
      account_id = "557058:aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
      role       = "admin"
    },
    {
      account_id = "557058:ffffffff-gggg-hhhh-iiii-jjjjjjjjjjjj"
      role       = "member"
    }
  ]
}

# Create a management team
resource "atlassian_team" "mgmt_team" {
  name        = "Management Team"
  description = "Executive and management team"
  type        = "management"

  members = [
    {
      account_id = "557058:zzzzzzzz-yyyy-xxxx-wwww-vvvvvvvvvvvv"
      role       = "admin"
    }
  ]
}
