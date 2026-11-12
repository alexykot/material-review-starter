variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "europe-west1"
}

variable "env" {
  type    = string
  default = "dev"
}

variable "api_image_tag" {
  description = "Tag of the API image in Artifact Registry (CI passes the git SHA)."
  type        = string
}

variable "agent_id" {
  description = "Managed Agent ID from `make agent-setup`."
  type        = string
  default     = ""
}

variable "environment_id" {
  description = "Managed Agents environment ID from `make agent-setup`."
  type        = string
  default     = ""
}
