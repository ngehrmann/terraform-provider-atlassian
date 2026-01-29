variable "atlassian_api_token" {
  description = "Atlassian API token for authentication"
  type        = string
  sensitive   = true
  default     = null
}

variable "atlassian_email" {
  description = "Email address associated with the Atlassian account"
  type        = string
  default     = null
}

variable "atlassian_organization" {
  description = "Atlassian organization/site name"
  type        = string
  default     = null
}

variable "atlassian_base_url" {
  description = "Base URL for Atlassian API"
  type        = string
  default     = "https://api.atlassian.com"
}

variable "atlassian_org_id" {
  description = "Atlassian organization ID"
  type        = string
  default     = null
}

variable "atlassian_site_id" {
  description = "Atlassian site ID"
  type        = string
  default     = null
}
