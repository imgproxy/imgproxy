#!/bin/bash
set -e

BASE_IMAGE="ghcr.io/imgproxy/imgproxy-base"
WORKFLOWS_DIR=".github/workflows"
DOCKERFILES="docker/Dockerfile"
UPDATED_WORKFLOWS=()

# Script to update base image versions across the project
# Usage: ./update-base-image.sh <new-version>
# Example: ./update-base-image.sh v4.2.0

if [ -z "$1" ]; then
    echo "Usage: $0 <new-version>"
    echo "Example: $0 v4.2.0"
    exit 1
fi

NEW_VERSION="$1"

# Validate version format (should start with v)
if [[ ! "$NEW_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format vX.Y.Z (e.g., v4.2.0)"
    exit 1
fi

echo "Updating base image versions to $NEW_VERSION..."

# Update Makefile
if [ -f "Makefile" ]; then
    sed -i.bak "s|\(BASE_IMAGE ?= $BASE_IMAGE:\).*|\1$NEW_VERSION|" Makefile
    rm -f Makefile.bak
    echo "✓ Updated Makefile"
fi

# Update GitHub Actions workflows
if [ -d "$WORKFLOWS_DIR" ]; then
    for workflow in $WORKFLOWS_DIR/*.yml $WORKFLOWS_DIR/*.yaml; do
        if [ -f "$workflow" ] && grep -q "$BASE_IMAGE" "$workflow"; then
            sed -i.bak "s|\(image: .*$BASE_IMAGE:\).*|\1$NEW_VERSION|g" "$workflow"
            rm -f "${workflow}.bak"
            echo "✓ Updated $workflow"
            UPDATED_WORKFLOWS+=("$workflow")
        fi
    done
fi

# Update .devcontainer/devcontainer.json (resolve symlink)
if [ -L ".devcontainer/devcontainer.json" ]; then
    DEVCONTAINER_FILE=$(readlink ".devcontainer/devcontainer.json")
    # Handle relative path
    if [[ "$DEVCONTAINER_FILE" != /* ]]; then
        DEVCONTAINER_FILE=".devcontainer/$DEVCONTAINER_FILE"
    fi
    sed -i.bak "s|\(\"image\": \".*$BASE_IMAGE:\).*\"|\1$NEW_VERSION\"|" "$DEVCONTAINER_FILE"
    rm -f "${DEVCONTAINER_FILE}.bak"
    echo "✓ Updated $DEVCONTAINER_FILE"
fi

# Update Dockerfiles
for dockerfile in $DOCKERFILES; do
    if [ -f "$dockerfile" ]; then
        sed -i.bak "s|\(ARG BASE_IMAGE_VERSION=\"\).*\"|\1$NEW_VERSION\"|" "$dockerfile"
        rm -f "${dockerfile}.bak"
        echo "✓ Updated $dockerfile"
    fi
done

echo ""
echo "Done! Base image version updated to $NEW_VERSION"
echo ""
echo "Files updated:"
echo "  - Makefile"
for workflow in "${UPDATED_WORKFLOWS[@]}"; do
    echo "  - $workflow"
done
echo "  - .devcontainer/devcontainer.json"
for dockerfile in $DOCKERFILES; do
    echo "  - $dockerfile"
done
echo ""
