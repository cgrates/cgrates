#! /usr/bin/env sh
./build.sh
go test $(glide novendor)
exit $?
