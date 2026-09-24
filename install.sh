#!/usr/bin/env bash
set -Eeuo pipefail

APP_DIR="${NZBHARBOR_DIR:-/opt/nzbharbor}"
IMAGE="${NZBHARBOR_IMAGE:-ghcr.io/kasundigital/nzbharbor}"
TAG="${NZBHARBOR_TAG:-latest}"
PORT="${NZBHARBOR_PORT:-6789}"

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker is required but was not found."
  echo "Install Docker first: https://docs.docker.com/engine/install/"
  exit 1
fi

if ! docker compose version >/dev/null 2>&1; then
  echo "Docker Compose v2 is required but was not found."
  exit 1
fi

if [ "$(id -u)" -ne 0 ]; then
  echo "Run this installer with sudo/root so it can use ${APP_DIR}."
  exit 1
fi

mkdir -p "${APP_DIR}/config" "${APP_DIR}/downloads"

cat > "${APP_DIR}/docker-compose.yml" <<EOF
services:
  nzbharbor:
    image: ${IMAGE}:${TAG}
    container_name: nzbharbor
    restart: unless-stopped
    init: true
    stop_grace_period: 30s
    ports:
      - "${PORT}:6789"
    environment:
      NZBHARBOR_CONFIG: /config
      NZBHARBOR_DOWNLOADS: /downloads
    volumes:
      - ${APP_DIR}/config:/config
      - ${APP_DIR}/downloads:/downloads
EOF

cd "${APP_DIR}"

echo "Pulling ${IMAGE}:${TAG} ..."
docker compose pull

echo "Starting/updating NZBHarbor ..."
docker compose up -d --remove-orphans

echo
echo "NZBHarbor is running."
echo "Web UI: http://SERVER-IP:${PORT}"
echo "Install directory: ${APP_DIR}"
echo "Config: ${APP_DIR}/config"
echo "Downloads: ${APP_DIR}/downloads"
echo
echo "Use the same curl command again any time to update."
