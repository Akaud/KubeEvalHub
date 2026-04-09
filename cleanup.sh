#!/usr/bin/env bash

set -e

echo "Deleting Kubernetes resources..."

kubectl delete -f agent.yaml || true

echo "Deleting namespace..."

kubectl delete namespace kubeevalhub-agent --ignore-not-found=true

echo "Stopping Docker Compose and removing volumes..."

docker compose down -v

echo "Cleanup complete."