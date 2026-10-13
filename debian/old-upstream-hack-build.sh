#!/usr/bin/bash

# The former build script for docker-buildx was removed in
# https://github.com/docker/buildx/pull/3588 in favor of an in-docker build.
# Here we re-introduce the old script to maintain our build as it was performed
# back then for now.

set -e

: "${DESTDIR=./bin/build}"
: "${PACKAGE=github.com/docker/buildx}"
: "${VERSION=$(./hack/git-meta version)}"
: "${REVISION=$(./hack/git-meta revision)}"

: "${CGO_ENABLED=0}"
: "${GO_PKG=github.com/docker/buildx}"
: "${GO_EXTRA_FLAGS=}"
: "${GO_LDFLAGS=-X ${GO_PKG}/version.Version=${VERSION} -X ${GO_PKG}/version.Revision=${REVISION} -X ${GO_PKG}/version.Package=${PACKAGE}}"
: "${GO_EXTRA_LDFLAGS=}"

set -x
CGO_ENABLED=$CGO_ENABLED go build -mod vendor -trimpath ${GO_EXTRA_FLAGS} -ldflags "${GO_LDFLAGS} ${GO_EXTRA_LDFLAGS}" -o "${DESTDIR}/docker-buildx" ./cmd/buildx
