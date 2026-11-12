terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }

  # The bucket must exist before `terraform init`; create it once by hand.
  backend "gcs" {
    prefix = "envs/dev"
    # bucket is passed at init: terraform init -backend-config="bucket=<state-bucket>"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

locals {
  name   = "mr-${var.env}"
  labels = { app = "material-review", env = var.env }
}

resource "google_project_service" "apis" {
  for_each = toset([
    "run.googleapis.com",
    "artifactregistry.googleapis.com",
    "secretmanager.googleapis.com",
    "storage.googleapis.com",
  ])
  service            = each.value
  disable_on_destroy = false
}

resource "google_artifact_registry_repository" "images" {
  repository_id = local.name
  location      = var.region
  format        = "DOCKER"
  labels        = local.labels
  depends_on    = [google_project_service.apis]
}

# Documents (submittals, agent outputs). The API reaches it through the GCS
# S3-interoperability endpoint with an HMAC key stored in Secret Manager.
resource "google_storage_bucket" "documents" {
  name                        = "${var.project_id}-${local.name}-documents"
  location                    = var.region
  uniform_bucket_level_access = true
  public_access_prevention    = "enforced"
  labels                      = local.labels
}

# Secret containers only. Values are added out of band
# (gcloud secrets versions add ...), never through Terraform state.
resource "google_secret_manager_secret" "secrets" {
  for_each  = toset(["anthropic-api-key", "database-url", "s3-access-key-id", "s3-secret-access-key"])
  secret_id = "${local.name}-${each.value}"
  labels    = local.labels

  replication {
    auto {}
  }
  depends_on = [google_project_service.apis]
}

module "api" {
  source = "../../modules/api"

  project_id = var.project_id
  region     = var.region
  name       = local.name
  image      = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.images.repository_id}/api:${var.api_image_tag}"
  labels     = local.labels

  env = {
    COOKIE_SECURE  = "true"
    S3_ENDPOINT    = "https://storage.googleapis.com"
    S3_REGION      = "auto"
    S3_BUCKET      = google_storage_bucket.documents.name
    AGENT_ID       = var.agent_id
    ENVIRONMENT_ID = var.environment_id
  }

  secret_env = {
    ANTHROPIC_API_KEY    = google_secret_manager_secret.secrets["anthropic-api-key"].secret_id
    DATABASE_URL         = google_secret_manager_secret.secrets["database-url"].secret_id
    S3_ACCESS_KEY_ID     = google_secret_manager_secret.secrets["s3-access-key-id"].secret_id
    S3_SECRET_ACCESS_KEY = google_secret_manager_secret.secrets["s3-secret-access-key"].secret_id
  }

  depends_on = [google_project_service.apis]
}
