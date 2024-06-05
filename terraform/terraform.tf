terraform {
  required_version = ">= 1.8.3"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.32.0"
    }
  }
}

provider "google" {
  project = var.google_cloud_project
  region  = "asia-northeast1"
  zone    = "asia-northeast1-a"
}

variable "google_cloud_project" {
  type        = string
  description = "Google Cloud Project ID"
}

resource "google_service_account" "functions_samples" {
  account_id = "functions-samples"
}

resource "google_project_iam_member" "functions_samples_pubsub_publisher" {
  project = var.google_cloud_project
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${google_service_account.functions_samples.email}"
}

resource "google_project_iam_member" "functions_samples_pubsub_subscriber" {
  project = var.google_cloud_project
  role    = "roles/pubsub.subscriber"
  member  = "serviceAccount:${google_service_account.functions_samples.email}"
}

resource "google_pubsub_topic" "functions_samples" {
  name = "functions-samples"
}
