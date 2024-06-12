#!/usr/bin/env bash

set -euxo pipefail

cd "$(dirname "$0")"

# you can create a bucket with the following command:
#
#   ```
#   $ gcloud storage buckets create "gs://${TFSTATE_BUCKET}"
#   ```
#
gcloud storage cp ./terraform.tfstate* "gs://${TFSTATE_BUCKET}"
