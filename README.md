# Terraform Provider for Atlassian

This Terraform provider enables you to manage Atlassian Teams using Terraform.

## Features

- Create, update, and delete Atlassian Teams
- Manage team members and their roles
- Full Terraform lifecycle support (Create, Read, Update, Delete, Import)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
```sh
git clone https://github.com/nikolas/terraform-provider-atlassian
cd terraform-provider-atlassian
```

2. Build the provider
```sh
make build
```

## Using the provider

### Authentication

The provider requires the following authentication parameters:

- **API Token**: Your Atlassian API token
- **Email**: Email address associated with your Atlassian account
- **Organization**: Your Atlassian organization/site name

These can be provided via:
1. Provider configuration block
2. Environment variables:
   - `ATLASSIAN_API_TOKEN`
   - `ATLASSIAN_EMAIL`
   - `ATLASSIAN_ORGANIZATION`
   - `ATLASSIAN_BASE_URL` (optional, defaults to https://api.atlassian.com)

### Provider Configuration

```hcl
terraform {
  required_providers {
    atlassian = {
      source  = "hashicorp/atlassian"
      version = "~> 0.1.0"
    }
  }
}

provider "atlassian" {
  api_token    = var.atlassian_api_token
  email        = var.atlassian_email
  organization = var.atlassian_organization
  # base_url   = "https://api.atlassian.com" # Optional
}
```

### Example Usage

#### Creating a Team

```hcl
resource "atlassian_team" "example" {
  name        = "Development Team"
  description = "Main development team for our products"
  type        = "development"
  
  members = [
    {
      account_id = "557058:12345678-1234-1234-1234-123456789012"
      role       = "admin"
    },
    {
      account_id = "557058:87654321-4321-4321-4321-210987654321"
      role       = "member"
    }
  ]
}
```

#### Importing an Existing Team

```sh
terraform import atlassian_team.example team_id_here
```

## Resource Reference

### `atlassian_team`

Manages an Atlassian Team.

#### Arguments

- `name` (Required) - The name of the team
- `description` (Optional) - Description of the team
- `type` (Required) - Type of team (e.g., "development", "support", "management")
- `organization` (Optional) - Organization identifier (defaults to provider organization)
- `members` (Optional) - Set of team members
  - `account_id` (Required) - Account ID of the team member
  - `email` (Optional, Computed) - Email address of the team member
  - `role` (Optional, Computed) - Role of the team member in the team

#### Attributes

- `id` - The unique identifier of the team
- `created_at` - Timestamp when the team was created
- `updated_at` - Timestamp when the team was last updated

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) 1.21+
- [Terraform](https://www.terraform.io/downloads.html) 1.0+
- [Make](https://www.gnu.org/software/make/)

### Building

```sh
make build
```

### Testing

Run unit tests:
```sh
make test
```

Run acceptance tests:
```sh
make testacc
```

### Installing Locally

To install the provider locally for development:

```sh
make install
```

This will build the provider and install it to your local Terraform plugins directory.

### Formatting

```sh
make fmt
```

### Documentation

Generate documentation:

```sh
make docs
```

## API Documentation

This provider interacts with Atlassian APIs. Please refer to the official Atlassian API documentation for more details:

- [Atlassian Admin API](https://developer.atlassian.com/cloud/admin/)
- [Atlassian Teams API](https://developer.atlassian.com/cloud/admin/teams/)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter any issues or have questions, please:

1. Check the [documentation](docs/)
2. Search existing [issues](https://github.com/nikolas/terraform-provider-atlassian/issues)
3. Create a new issue if needed

## Changelog

### v0.1.0

- Initial release
- Support for managing Atlassian Teams
- CRUD operations for teams
- Team member management
- Import support