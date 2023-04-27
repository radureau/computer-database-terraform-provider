data "google_storage_bucket" "functions" {
  name = var.bucket_name
}

data "archive_file" "local_func_archive" {
  type        = "zip"
  source_dir  = var.function_source
  output_path = "${path.root}/build/cloudfunction/${var.function_name}.zip"
}

resource "google_storage_bucket_object" "remote_func_archive" {
  name   = format("%s-%s.zip", "cf_${var.function_name}", data.archive_file.local_func_archive.output_base64sha256)
  bucket = data.google_storage_bucket.functions.name
  source = data.archive_file.local_func_archive.output_path
}

resource "google_cloudfunctions_function" "function" {
  name                        = var.function_name
  description                 = var.function_description
  region                      = "europe-west1"
  runtime                     = var.function_runtime
  available_memory_mb         = var.memory
  source_archive_bucket       = data.google_storage_bucket.functions.name
  source_archive_object       = google_storage_bucket_object.remote_func_archive.name
  trigger_http                = var.trigger_http
  entry_point                 = var.function_entrypoint
  max_instances               = var.max_instance
  ingress_settings            = var.ingress_settings
  timeout                     = var.timeout
  labels                      = var.labels
  service_account_email       = var.service_account_email
  environment_variables       = var.environment_variables
  build_environment_variables = var.build_environment_variables

  dynamic "secret_environment_variables" {
    for_each = var.secret_environment_variables
    iterator = each
    content {
      key        = each.value.key
      project_id = each.value.project_id
      secret     = each.value.secret
      version    = each.value.version
    }
  }
  dynamic "secret_volumes" {
    for_each = var.secret_volumes
    iterator = each
    content {
      mount_path = each.value.mount_path
      project_id = each.value.project_id
      secret     = each.value.secret
      dynamic "versions" {
        for_each = each.value.versions
        content {
          path = versions.value.path
          version = versions.value.version
        }
      }
    }
  }
}

resource "google_cloudfunctions_function_iam_member" "invoker" {
  project        = google_cloudfunctions_function.function.project
  region         = google_cloudfunctions_function.function.region
  cloud_function = google_cloudfunctions_function.function.name
  role           = "roles/cloudfunctions.invoker"

  for_each = toset(var.invokers)
  member   = each.key
}
