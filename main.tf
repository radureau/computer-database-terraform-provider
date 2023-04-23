terraform {
  required_providers {
    computer-database = {
      source  = "computer-database"
      version = ">=0.0.1"
    }
  }
}

provider "computer-database" {
  api_url = "http://localhost:8080/api/v1"
}

locals {
  computer_models = [
    "First",
    "Second",
    "Third",
  ]
}

resource "computer-database_company" "my_company" {
  id = "cotf"
  name = "My Terraformed Company"

  computer_models = toset([ for model_name in local.computer_models:
    {
      id = format("cotf%s", lower(model_name))
      name = model_name
      release = 2023
    }
  ])
}