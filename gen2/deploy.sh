#!/usr/bin/env bash

set -euxo pipefail

PUBSUB_TOPIC_ID="functions-samples"

# =========================================================
# コマンド引数に応じた関数設定の切り替え
# 例:
# $ ./gen2/deploy.sh start
# =========================================================
cd "$(dirname "$0")"

case "$1" in
  start)
    FUNC_NAME="functions-samples-start"
    TRIGGER_FLAGS="--trigger-http"
    ;;
  hook)
    FUNC_NAME="functions-samples-hook"
    TRIGGER_FLAGS="--trigger-topic ${PUBSUB_TOPIC_ID}"
    ;;
  *)
    echo "Usage: $0 {start|hook}"
    exit 1
    ;;
esac

gcloud functions deploy "${FUNC_NAME}" \
  --project "${GOOGLE_CLOUD_PROJECT}" \
  --gen2 \
  --region asia-northeast1 \
  --runtime go122 \
  --run-service-account "functions-samples@${GOOGLE_CLOUD_PROJECT}.iam.gserviceaccount.com" \
  --source . \
  --entry-point "${FUNC_NAME}" \
  ${TRIGGER_FLAGS} \
  --quiet \
  --set-env-vars "GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT}" \
  --set-env-vars "PUBSUB_TOPIC_ID=${PUBSUB_TOPIC_ID}"
