# Atlassian Provider Configuration

## Authentication & Configuration Parameters

The Atlassian provider requires several configuration parameters to properly authenticate and route API requests to the correct Atlassian services.

### Required Parameters

- `api_token`: Your Atlassian API token for authentication
- `email`: Email address associated with your Atlassian account

### Organization & Site Identification

Atlassian has different types of identifiers for different API endpoints:

#### `organization` (String, Optional)
- **Purpose**: Legacy organization name or identifier
- **Used for**: Fallback when `org_id` is not provided
- **Format**: Human-readable organization name
- **Example**: `"my-company"`

#### `org_id` (String, Optional)  
- **Purpose**: Organization ID for Atlassian Admin APIs
- **Used for**: Team management, user management, organization settings
- **Format**: UUID or organization-specific identifier
- **Example**: `"12345678-1234-1234-1234-123456789012"`
- **APIs**: Admin API endpoints (`/admin/v1/orgs/{org_id}/...`)

#### `site_id` (String, Optional)
- **Purpose**: Site ID (cloudid) for product-specific APIs
- **Used for**: Jira, Confluence, and other Atlassian product APIs
- **Format**: UUID representing the specific Atlassian site/instance
- **Example**: `"abcd1234-5678-90ab-cdef-1234567890ab"`
- **APIs**: Jira API (`/ex/jira/{site_id}/...`), Confluence API, etc.

### When to Use Each ID

| API Type | Required ID | Example Endpoint | Use Case |
|----------|-------------|-----------------|----------|
| Admin APIs (Teams, Users, Org Settings) | `org_id` or `organization` | `/admin/v1/orgs/{org_id}/teams` | Managing teams and organization |
| Jira APIs | `site_id` | `/ex/jira/{site_id}/rest/api/3/issue` | Creating/managing Jira issues |
| Confluence APIs | `site_id` | `/wiki/api/v2/spaces` | Managing Confluence spaces/pages |

### Environment Variables

All parameters can be set via environment variables:

- `ATLASSIAN_API_TOKEN`
- `ATLASSIAN_EMAIL` 
- `ATLASSIAN_ORGANIZATION`
- `ATLASSIAN_ORG_ID`
- `ATLASSIAN_SITE_ID`
- `ATLASSIAN_BASE_URL`

### Finding Your IDs

#### Finding your Site ID (cloudid)
1. Go to your Atlassian site (e.g., `https://your-domain.atlassian.net`)
2. The site ID is in the URL or can be obtained via: `https://your-domain.atlassian.net/_edge/tenant_info`

#### Finding your Organization ID
1. Access Atlassian Admin console
2. Check the URL or use the Admin API to retrieve organization details

### Example Configuration

```terraform
provider "atlassian" {
  api_token    = var.atlassian_api_token
  email        = "admin@company.com"
  org_id       = "12345678-1234-1234-1234-123456789012"  # For admin operations
  site_id      = "abcd1234-5678-90ab-cdef-1234567890ab"  # For Jira/Confluence
  base_url     = "https://api.atlassian.com"              # Optional, defaults to this
}
```