#!/bin/bash
set -e

# Default values
BUILD_DIR_DEFAULT=$HOME/cgr_build
SRCDIR_DEFAULT=$HOME/go/src/github.com/cgrates/cgrates

# Parse options
while (( "$#" )); do
  case "$1" in
    --build-dir)
      BUILD_DIR=$2
      shift 2
      ;;
    --srcdir)
      SRCDIR=$2
      shift 2
      ;;
    -*|--*=) # unsupported flags
      echo "Error: Unsupported flag $1" >&2
      exit 1
      ;;
    *) # preserve positional arguments
      PARAMS="$PARAMS $1"
      shift
      ;;
  esac
done
# set positional arguments in their proper place
eval set -- "$PARAMS"

# Assign defaults if variables are not set
BUILD_DIR=${BUILD_DIR:-$BUILD_DIR_DEFAULT}
SRCDIR=${SRCDIR:-$SRCDIR_DEFAULT}

prepare_environment() {
    echo "Making sure dependencies are installed..."
    sudo dnf install -y rpm-build wget curl tar
    echo "Creating build directories in $BUILD_DIR..."
    mkdir -p $BUILD_DIR/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
}

fetch_source() {
    echo "Fetching source code..."
    cd $SRCDIR
    export gitLastCommit=$(git rev-parse HEAD)
    export rpmTag=$(git log -1 --format=%ci | date +%Y%m%d%H%M%S)+$(git rev-parse --short HEAD)
    if [ ! -f $BUILD_DIR/SOURCES/$gitLastCommit.tar.gz ]; then
        wget -P $BUILD_DIR/SOURCES https://github.com/cgrates/cgrates/archive/$gitLastCommit.tar.gz
    fi
}

copy_spec_file() {
    echo "Copying RPM spec file..."
    cp $SRCDIR/packages/redhat_fedora/cgrates.spec $BUILD_DIR/SPECS
}

build_package() {
    echo "Building RPM package..."
    cd $BUILD_DIR
    rpmbuild -bb --define "_topdir $BUILD_DIR" SPECS/cgrates.spec
}

main() {
    prepare_environment
    fetch_source
    copy_spec_file
    build_package
}

main "$@"