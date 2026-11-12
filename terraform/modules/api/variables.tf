variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "name" {
  description = "Prefix for resource names, e.g. \"mr-dev\"."
  type        = string
}

variable "image" {
  description = "Full Artifact Registry image reference for the API."
  type        = string
}

variable "env" {
  description = "Plain environment variables."
  type        = map(string)
  default     = {}
}

variable "secret_env" {
  description = "Environment variable name => Secret Manager secret ID."
  type        = map(string)
  default     = {}
}

variable "min_instances" {
  type    = number
  default = 0
}

variable "max_instances" {
  type    = number
  default = 3
}

variable "labels" {
  type    = map(string)
  default = {}
}
