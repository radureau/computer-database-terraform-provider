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
  name = "My Terraformed Company 2"

  computer_model {
    id = "cotfcmtf"
    name = "My Terraformed Computer Model"
    release = 2023
  }

  dynamic "computer_model" {
    for_each = toset(local.computer_models)
    content {
      id = format("cotf%s", lower(computer_model.value))
      name = computer_model.value
      release = 2023
    }
  }
}