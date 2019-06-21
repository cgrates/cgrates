#! /usr/bin/env sh

go test github.com/cgrates/cgrates/config -tags=generate | echo "Generating configuration file ..."
