#!/usr/bin/env bash
# Builds the provider and installs it into the local Terraform plugin mirror.
#
#   ./build.sh               installs as the release version in main.go
#   ./build.sh 1.1.0-dev     installs as a separate local test version
#
# A test version is injected at build time with -X main.version, the same way GoReleaser
# stamps releases, so main.go is never edited. Keep it plain SemVer with a pre-release suffix
# (1.1.0-dev, 1.1.0-dev.2): BioT validates the version at provider start-up, and its parser
# fails on build metadata such as 1.1.0+local.
#
# Terraform only selects a pre-release version when a configuration pins it exactly
# (version = "= 1.1.0-dev"), so a test build can't be picked up by other projects by accident.
set -euo pipefail

RELEASE_VERSION=$(grep 'version string' main.go | grep -o '"[^"]*"' | tr -d '"')
VERSION="${1:-$RELEASE_VERSION}"

PLATFORM="$(go env GOOS)_$(go env GOARCH)"
PLUGIN_DIR=~/.terraform.d/plugins/registry.terraform.io/biot-med/biot-gen2/${VERSION}/${PLATFORM}

go build -ldflags "-X main.version=${VERSION}" -o terraform-provider-biot-gen2
mkdir -p "${PLUGIN_DIR}"
cp terraform-provider-biot-gen2 "${PLUGIN_DIR}/"

echo "Installed biot-gen2 ${VERSION} (${PLATFORM}) to ${PLUGIN_DIR}"
if [ "${VERSION}" != "${RELEASE_VERSION}" ]; then
  echo "Pin it in your test project with:  version = \"= ${VERSION}\""
fi
