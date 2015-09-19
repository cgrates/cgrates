#! /usr/bin/env sh

echo "Disabling Glide..."

echo "export GO15VENDOREXPERIMENT=0"
export GO15VENDOREXPERIMENT=0

echo "rm -rf vendor"
rm -rf vendor

go get -v -u github.com/cgrates/cgrates/...
gg=$?

exit $gg
