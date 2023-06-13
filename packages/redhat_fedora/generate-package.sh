#!/bin/bash
set -e

WORKDIR=$HOME/cgr_build
SRCDIR=$HOME/go/src/github.com/cgrates/cgrates

prepare_environment() {
    echo "Preparing environment..."
    sudo dnf install -y rpm-build
    mkdir -p $WORKDIR/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
}

fetch_source() {
    echo "Fetching source code..."
    cd $SRCDIR
    gitLastCommit=$(git rev-parse HEAD)
    rpmTag=$(git log -1 --format=%ci | date +%Y%m%d%H%M%S)+$(git rev-parse --short HEAD)
    wget -P $WORKDIR/SOURCES https://github.com/cgrates/cgrates/archive/$gitLastCommit.tar.gz
}

copy_spec_file() {
    echo "Copying spec file..."
    cp $SRCDIR/packages/redhat_fedora/cgrates.spec $WORKDIR/SPECS
}

build_package() {
    echo "Building RPM package..."
    cd $WORKDIR
    rpmbuild -bb --define "_topdir $WORKDIR" SPECS/cgrates.spec
}

cleanup() {
    echo "Cleaning up..."
    rm -rf $WORKDIR/SOURCES/*
}

main() {
    prepare_environment
    fetch_source
    copy_spec_file
    build_package
    cleanup
}

main "$@"

