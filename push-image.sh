#!/bin/bash

# Script to build and push Docker image to GitHub Container Registry
set -e

GITHUB_USER="timcsy"
IMAGE_NAME="domain-manager"
VERSION="${1:-latest}"

echo "🔨 Building Docker image..."
docker build -t ${IMAGE_NAME}:${VERSION} -f Dockerfile .

echo "🏷️  Tagging image for GitHub Container Registry..."
docker tag ${IMAGE_NAME}:${VERSION} ghcr.io/${GITHUB_USER}/${IMAGE_NAME}:${VERSION}

echo "📤 Pushing image to GitHub Container Registry..."
echo "⚠️  Make sure you've logged in to GitHub Container Registry first:"
echo "   docker login ghcr.io -u ${GITHUB_USER}"
echo ""
read -p "Press Enter to continue with push, or Ctrl+C to cancel..."

docker push ghcr.io/${GITHUB_USER}/${IMAGE_NAME}:${VERSION}

echo "✅ Image pushed successfully!"
echo "📦 Image: ghcr.io/${GITHUB_USER}/${IMAGE_NAME}:${VERSION}"
