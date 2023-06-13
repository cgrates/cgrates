#!/bin/bash
./build.sh
go test ./...
exit $?
