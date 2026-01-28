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
