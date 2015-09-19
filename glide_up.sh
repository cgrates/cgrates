#! /usr/bin/env sh

echo "Executing Glide..."

go get -v github.com/Masterminds/glide
gl=$?
glide up
gu=$?

exit $gl || $gu

echo "export GO15VENDOREXPERIMENT=1"
export GO15VENDOREXPERIMENT=1
