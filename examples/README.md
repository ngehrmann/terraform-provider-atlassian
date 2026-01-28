# Terraform Atlassian Provider Examples

This directory contains example Terraform configurations that demonstrate how to use the Atlassian provider.

## Prerequisites

1. **Atlassian Account**: You need an Atlassian account with appropriate permissions
2. **API Token**: Generate an API token from [Atlassian Account Settings](https://id.atlassian.com/manage-profile/security/api-tokens)
3. **Organization Access**: Ensure you have admin access to your Atlassian organization

## Setup

1. **Copy the example variables file**:
   ```sh
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Edit terraform.tfvars** with your actual values:
   ```hcl
   atlassian_api_token    = "your-actual-api-token"
   atlassian_email        = "your-email@company.com"
   atlassian_organization = "your-org-name"
   ```

3. **Alternative: Use environment variables**:
   ```sh
   export ATLASSIAN_API_TOKEN="your-actual-api-token"
   export ATLASSIAN_EMAIL="your-email@company.com"
   export ATLASSIAN_ORGANIZATION="your-org-name"
   ```

## Running the Example

1. **Initialize Terraform**:
   ```sh
   terraform init
   ```

2. **Plan the deployment**:
   ```sh
   terraform plan
   ```

3. **Apply the configuration**:
   ```sh
   terraform apply
   ```

4. **View outputs**:
   ```sh
   terraform output
   ```

5. **Clean up when done**:
   ```sh
   terraform destroy
   ```

## What This Example Creates

This example creates three teams:

1. **Development Team**: A team for developers with 3 members
2. **Customer Support**: A team for support staff with 2 members
3. **Management Team**: A team for management with 1 member

## Customizing the Example

### Adding More Teams

You can add more teams by creating additional `atlassian_team` resources:

```hcl
resource "atlassian_team" "qa_team" {
  name        = "Quality Assurance"
  description = "QA and testing team"
  type        = "qa"
  
  members = [
    {
      account_id = "your-account-id"
      role       = "admin"
    }
  ]
}
```

### Managing Team Members

To find account IDs for team members:

1. Go to your Atlassian Admin console
2. Navigate to Users
3. Find the user and copy their Account ID

The Account ID format is typically: `557058:xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

### Team Types

Common team types include:
- `development`
- `support`
- `management`
- `qa`
- `operations`
- `marketing`
- `sales`

You can use any string value that makes sense for your organization.

## Troubleshooting

### Authentication Issues

- Verify your API token is correct and hasn't expired
- Ensure your email matches the account that created the API token
- Check that your organization name is correct

### Permission Issues

- Ensure you have admin permissions in your Atlassian organization
- Verify that the API token has the necessary scopes

### Team Creation Issues

- Check that account IDs for members are valid and exist in your organization
- Ensure team names are unique within your organization

## Next Steps

- Review the [main README](../README.md) for more detailed documentation
- Check the [Atlassian API documentation](https://developer.atlassian.com/cloud/admin/) for more advanced use cases
- Consider creating modules to standardize team creation across your organization