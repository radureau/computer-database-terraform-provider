variable "bucket_name" {
  type        = string
  nullable    = false
  description = "bucket to store cloud function source code"
}
variable "function_name" {
  type     = string
  nullable = false
}
variable "function_entrypoint" {
  type     = string
  nullable = false
}
variable "function_source" {
  type     = string
  nullable = false
}
variable "function_description" {
  type    = string
  default = "no description"
}
variable "function_runtime" {
  type    = string
  default = "python39"
}
variable "invokers" {
  type    = list(string)
  default = []
}

variable "memory" {
  type    = number
  default = 128
}

variable "trigger_http" {
  type    = bool
  default = true
}

variable "timeout" {
  type    = number
  default = 60
}

variable "ingress_settings" {
  type    = string
  default = "ALLOW_INTERNAL_ONLY"
}

variable "labels" {
  type    = map(any)
  default = {}
}

variable "max_instance" {
  type    = number
  default = 10
}

variable "service_account_email" {
  type    = string
  default = null
}

variable "environment_variables" {
  description = "Cloud function environment variables."
  type        = map(string)
  default     = {}
}
variable "build_environment_variables" {
  type    = map(string)
  default = {}
}
variable "secret_environment_variables" {
  type = list(object({
    key        = string
    project_id = optional(string)
    secret     = string
    version    = string
  }))
  default = []
} #https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloudfunctions_function#nested_secret_environment_variables
variable "secret_volumes" {
  type = list(object({
    mount_path = string
    project_id = optional(string)
    secret     = string
    versions = optional(list(object({
      path    = string
      version = string
    })))
  }))
  default = []
} #https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloudfunctions_function#nested_secret_volumes
