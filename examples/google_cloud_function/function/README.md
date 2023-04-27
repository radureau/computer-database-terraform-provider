# USAGE

```
resource "google_storage_bucket" "functions" {
  name     = "oxyl-tn-2023-04-functions"
  location = "europe-west1"
}

module "cf_helloworld" {
  depends_on = [
    google_storage_bucket.functions,
  ]
  source = "<path to function blueprint folder>"
  bucket_name = google_storage_bucket.functions.name
  function_name = "helloworld"
  function_entrypoint = "hello_world"
  function_source = "${path.root}/src/cloudfunction/${var.function_name}/"
  environment_variables = {
    SQL_USER = "user"
  }
}
```

# ATTRIBUTE REFERENCES

[resources/cloudfunctions_function](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloudfunctions_function#argument-reference)

# CONTRIBUTION GUIDELINES

Make an Issue and either assign @radurga or ask for review if you can provide the code.
Don't forget to set the right Milestone.
Ping on Teams to make sure your need is heard.

[google provided blueprint](https://github.com/GoogleCloudPlatform/cloud-foundation-fabric/tree/master/modules/cloud-function) for reference