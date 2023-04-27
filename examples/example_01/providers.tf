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