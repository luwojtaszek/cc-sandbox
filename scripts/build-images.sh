#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
IMAGE_PREFIX="${IMAGE_PREFIX:-cc-sandbox}"

cd "$PROJECT_DIR"

echo "🔨 Building cc-sandbox Docker images..."

echo "📦 Building $IMAGE_PREFIX:base"
docker build -t "$IMAGE_PREFIX:base" ./docker/base

echo "📦 Building $IMAGE_PREFIX:docker"
docker build -t "$IMAGE_PREFIX:docker" ./docker/docker

echo "📦 Building $IMAGE_PREFIX:bun-full"
docker build -t "$IMAGE_PREFIX:bun-full" ./docker/bun-full

echo "✅ All images built successfully!"
docker images --filter "reference=$IMAGE_PREFIX:*" --format "table {{.Repository}}:{{.Tag}}\t{{.Size}}"
