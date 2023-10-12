#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $0 <ORIGTARGZ>"
    exit 1
fi

if [ -z "${GO}" ]; then
    GO="/usr/local/go/bin/go"
fi

UPSTREAM_TARBALL="$(realpath -s "$1")"

if [ ! -e "${UPSTREAM_TARBALL}" ]; then
    echo "Error: Upstream tarball not found"
    exit 1
fi

COMPONENT_NAME="dependencies"
COMPONENT_TARBALL="${UPSTREAM_TARBALL//.orig.tar/.orig-${COMPONENT_NAME}.tar}"

TEMP_DIR="$(mktemp -d)"

GOPATH="${TEMP_DIR}/${COMPONENT_NAME}"
export GOPATH

echo "Unpacking upstream tarball: ${UPSTREAM_TARBALL} into: ${TEMP_DIR}"
tar --strip-components=1 -xaf "${UPSTREAM_TARBALL}" -C "${TEMP_DIR}"

for CMD_DIR in "${TEMP_DIR}/cmd/"*; do
    echo "Getting $(basename "${CMD_DIR}") dependencies into: ${GOPATH}"
    cd "${CMD_DIR}" || exit 1
    "${GO}" get .
    cd "${OLDPWD}" || exit 1
done

echo "Fixing permissions for: ${GOPATH}"
chmod -R u+w "${GOPATH}"

echo "Creating component tarball: ${COMPONENT_TARBALL}"
cd "${TEMP_DIR}" || exit 1
tar --owner root --group root -caf "${COMPONENT_TARBALL}" "${COMPONENT_NAME}"
cd "${OLDPWD}" || exit 1

echo "Removing temporary directory: ${TEMP_DIR}"
rm -rf "${TEMP_DIR}"
