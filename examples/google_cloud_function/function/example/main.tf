provider "google" {
  project     = "oxyl-terraform-tn-april-23"
  region      = "europe-west1"
}

resource "google_storage_bucket" "functions" {
  name     = "oxyl-tn-2023-04-functions"
  location = "europe-west1"
}

module "cf_helloworld" {
  depends_on = [
    google_storage_bucket.functions,
  ]
  source              = "../../function"
  bucket_name         = google_storage_bucket.functions.name
  function_name       = "helloworld"
  function_entrypoint = "hello_world"
  function_source     = "${path.root}/src/cloudfunction/helloworld/"
  environment_variables = {
    SQL_USER = "user"
  }
}