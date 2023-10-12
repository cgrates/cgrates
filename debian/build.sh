#!/bin/bash

################################################################################
# Dependencies

PACKAGES=(
    git
    distro-info
    dpkg-dev
    devscripts
    pbuilder
    cowbuilder
    debhelper
    dh-golang
)

MISSING=()
for PKG in "${PACKAGES[@]}"; do
    INSTALLED="$(dpkg-query -W --showformat='${Status}' "${PKG}" | grep -c " ok installed")"
    if [ "${INSTALLED}" != "1" ]; then
        MISSING+=("${PKG}")
    fi
done

if [ "${#MISSING[@]}" != "0" ]; then
    echo "Error: Not all dependencies are installed: ${MISSING[*]}"
    exit 1
fi

################################################################################
# Variables

DEFAULT_DISTRIBUTION="bookworm"

if [ -z "${DISTRIBUTION}" ]; then
    DISTRIBUTION="${DEFAULT_DISTRIBUTION}"
fi

DEFAULT_RELEASE_VERSION="$(distro-info -r --series "${DISTRIBUTION}" 2> /dev/null)"

if [ -z "${RELEASE_VERSION}" ]; then
    RELEASE_VERSION="${DEFAULT_RELEASE_VERSION}"
fi

DEFAULT_CHROOT="/var/cache/pbuilder/base-${DISTRIBUTION}+go.cow"

if [ -z "${CHROOT}" ]; then
    CHROOT="${DEFAULT_CHROOT}"
fi

DEFAULT_NOCHROOT="0"

if [ -z "${NOCHROOT}" ]; then
    NOCHROOT="${DEFAULT_NOCHROOT}"
fi

DEFAULT_DEBUG="0"

if [ -z "${DEBUG}" ]; then
    DEBUG="${DEFAULT_DEBUG}"
fi

################################################################################
# Commandline options

OPTS=$(getopt -o D:R:C:Ndh --long distribution:release-version:,chroot:,nochroot,debug,help -n "$(basename "$0")" -- "$@")
RC=$?

if [ "${RC}" != 0 ]; then
    echo "Error: Failed to parse options."
    exit 1
fi

eval set -- "${OPTS}"

while true; do
    case "$1" in
        -D|--distribution)
            shift
            DISTRIBUTION="$1"
            shift
            ;;
        -R|--release-version)
            shift
            RELEASE_VERSION="$1"
            shift
            ;;
        -C|--chroot)
            shift
            CHROOT="$1"
            shift
            ;;
        -N|--nochroot)
            shift
            DEBUG=1
            ;;
        -d|--debug)
            shift
            DEBUG=1
            ;;
        -h|--help)
            shift

            echo "Usage: $(basename "$0") [OPTIONS]"
            echo
            echo "Options:"
            echo "-D, --distribution <NAME>        Distribution to use in changelog"
            echo "                                 Default: ${DEFAULT_DISTRIBUTION}"
            echo "-R, --release-version <NUMBER>   Version to use in ~deb<N>u1 suffix"
            echo "                                 Default: ${DEFAULT_RELEASE_VERSION}"
            echo "                                 Set to 0 to not append the suffix"
            echo "-C, --chroot <PATH>              Path to cowbuilder chroot"
            echo "                                 Default: ${DEFAULT_CHROOT}"
            echo "-N, --nochroot                   Don't use chroot for package build"
            echo "-d, --debug                      Enable debug output"
            echo "-h, --help                       Display this usage information"
            echo
            echo "Environment variables:"
            echo "DISTRIBUTION                     Distribution to use in changelog"
            echo "RELEASE_VERSION                  Version to use in ~deb<N>u1 suffix"
            echo "                                 Set to 0 to not append the suffix"
            echo "CHROOT                           Path to cowbuilder chroot"
            echo "NOCHROOT                         Don't use chroot for package build"
            echo "DEBUG                            Enable debug output"
            echo "                                 Set to 1 to enable debug output"

            exit 1
            ;;
        --)
            shift
            break
            ;;
        *)
            shift
            break
            ;;
    esac
done

################################################################################
# Main

if [ "${DEBUG}" = "1" ]; then
    set -x
fi

#
# Create .orig.tar.gz
#

DEBIAN_DIR="$(dirname "$0")"
SOURCE_DIR="$(dirname "${DEBIAN_DIR}")"

cd "${SOURCE_DIR}" || exit 1

PACKAGE="$(dpkg-parsechangelog -S Source)"
VERSION="$(grep -E "(^|\s+)Version\s*=\s*\"(\S+)\"\s*" utils/consts.go | awk -F'"' '{print $2}' | sed 's/^v//g')"

PATTERN="^[0-9]+.[0-9]+.[0-9]+(~[a-z0-9]+)?$"

if [[ ${VERSION} =~ ${PATTERN} ]]; then
    true
else
    echo "Error: Failed to extract version"
    exit 1
fi

GIT_COMMIT="HEAD"

GIT_TAG_LOG="$(git tag -l --points-at "${GIT_COMMIT}")"

COMMIT_DATE="$(git log -n1 --format=format:%cd --date="format:%Y%m%d%H%M%S" "${GIT_COMMIT}")"
COMMIT_HASH="$(git log -n1 --format=format:%h "${GIT_COMMIT}")"

if [ -n "${GIT_TAG_LOG}" ]; then
    PACKAGE_VERSION="${VERSION}"
else
    PACKAGE_VERSION="${VERSION}+${COMMIT_DATE}+${COMMIT_HASH}"
fi

ORIG_TARBALL="../${PACKAGE}_${PACKAGE_VERSION}.orig.tar.gz"

if [ ! -e "${ORIG_TARBALL}" ]; then
    echo "Creating ${ORIG_TARBALL} from ${GIT_COMMIT}"
    git archive -o "${ORIG_TARBALL}" --format tar.gz --prefix "${PACKAGE}-${VERSION}/" "${GIT_COMMIT}"
fi

#
# Create .orig.tar-dependencies.gz
#

DEPENDENCIES_TARBALL="../${PACKAGE}_${PACKAGE_VERSION}.orig-dependencies.tar.gz"

if [ ! -e "${DEPENDENCIES_TARBALL}" ]; then
    ./debian/create-components.sh "${ORIG_TARBALL}"
fi

#
# Unpack .orig.tar-dependencies.gz
#

if [ ! -e "${DEPENDENCIES_TARBALL}" ]; then
    echo "Error: No dependencies tarball"
    exit 1
fi

DEPENDENCIES_PATH="dependencies"

if [ ! -e "${DEPENDENCIES_PATH}" ]; then
    echo "Unpacking dependencies"
    tar xaf "${DEPENDENCIES_TARBALL}"
fi

#
# Update changelog
#

CHANGELOG_VERSION="$(dpkg-parsechangelog -S Version)"

PACKAGE_REVISION="${PACKAGE_VERSION}-1"

if [[ ${RELEASE_VERSION} =~ ^[0-9]+$ ]] && [ "${RELEASE_VERSION}" != "0" ]; then
    PACKAGE_REVISION="${PACKAGE_REVISION}~deb${RELEASE_VERSION}u1"
fi

if [ "${CHANGELOG_VERSION}" != "${PACKAGE_REVISION}" ]; then
    echo "Updating changelog"
    dch -v "${PACKAGE_REVISION}" \
        -m "Package build for git commit ${COMMIT_HASH} (${COMMIT_DATE})." \
        -D "${DISTRIBUTION}" --force-distribution
fi

#
# Build package
#

if [ -n "${GIT_TAG_LOG}" ]; then
    GIT_COMMIT_DATE=""
    GIT_COMMIT_HASH=""
else
    GIT_COMMIT_DATE="$(git log -n1 --format=format:%cI "${GIT_COMMIT}")"
    GIT_COMMIT_HASH="$(git log -n1 --format=format:%H "${GIT_COMMIT}")"
fi

export GIT_COMMIT_DATE
export GIT_COMMIT_HASH

echo "Building package"
if [ "${NOCHROOT}" = "0" ]; then
    pdebuild --pbuilder cowbuilder -- --basepath "${CHROOT}"
else
    MAINTAINER_EMAIL="$(dpkg-parsechangelog -S Maintainer | awk -F'<' '{print $2}' | sed 's/>$//')"
    KEY_COUNT="$(gpg --list-secret-keys "${MAINTAINER_EMAIL}" 2> /dev/null | grep -c "^sec")"

    if [ "${KEY_COUNT}" = "0" ]; then
        NO_SIGN="--no-sign"
    else
        NO_SIGN=""
    fi

    dpkg-buildpackage -rfakeroot -tc "${NO_SIGN}"
fi

#
# Undo changes
#

echo "Undoing changes"
rm -rf dependencies/
git checkout debian/changelog

