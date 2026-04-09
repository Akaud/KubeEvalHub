#!/usr/bin/env bash
set -Eeuo pipefail

IMAGE_NAME="kubeevalhub-agent:latest"
AGENT_DIR="./agent"

command -v docker >/dev/null 2>&1 || { echo "docker not found"; exit 1; }
command -v minikube >/dev/null 2>&1 || { echo "minikube not found"; exit 1; }

if [ ! -d "${AGENT_DIR}" ]; then
  echo "Agent directory not found: ${AGENT_DIR}"
  exit 1
fi

echo "[1/5] Stopping compose stack"
docker compose down || true

echo "[2/5] Removing old image"
docker image inspect "${IMAGE_NAME}" >/dev/null 2>&1 && docker image rm -f "${IMAGE_NAME}" || true

echo "[3/5] Building agent image"
docker build --no-cache -t "${IMAGE_NAME}" "${AGENT_DIR}"

echo "[4/5] Loading image into minikube"
minikube image load "${IMAGE_NAME}"

echo "[5/5] Starting compose stack"
docker compose up --build -d 