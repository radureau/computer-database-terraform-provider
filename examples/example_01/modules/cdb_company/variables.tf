variable "name" {
    type = string
    description = "company name"
}

variable "computer_models" {
    type = list(object({
        name = string
        release = optional(number)
    }))
    description = "set of computer models"
    default = []
}