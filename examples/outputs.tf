output "dev_team_id" {
  description = "ID of the development team"
  value       = atlassian_team.dev_team.id
}

output "dev_team_created_at" {
  description = "Creation timestamp of the development team"
  value       = atlassian_team.dev_team.created_at
}

output "support_team_id" {
  description = "ID of the support team"
  value       = atlassian_team.support_team.id
}

output "mgmt_team_id" {
  description = "ID of the management team"
  value       = atlassian_team.mgmt_team.id
}

output "all_teams" {
  description = "Summary of all created teams"
  value = {
    development = {
      id          = atlassian_team.dev_team.id
      name        = atlassian_team.dev_team.name
      description = atlassian_team.dev_team.description
      members     = length(atlassian_team.dev_team.members)
    }
    support = {
      id          = atlassian_team.support_team.id
      name        = atlassian_team.support_team.name
      description = atlassian_team.support_team.description
      members     = length(atlassian_team.support_team.members)
    }
    management = {
      id          = atlassian_team.mgmt_team.id
      name        = atlassian_team.mgmt_team.name
      description = atlassian_team.mgmt_team.description
      members     = length(atlassian_team.mgmt_team.members)
    }
  }
}
