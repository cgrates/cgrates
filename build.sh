#! /usr/bin/env sh

echo "Building CGRateS..."

go install github.com/cgrates/cgrates/cmd/cgr-engine
cr=$?
go install github.com/cgrates/cgrates/cmd/cgr-loader
cl=$?
go install github.com/cgrates/cgrates/cmd/cgr-console
cc=$?
go install github.com/cgrates/cgrates/cmd/cgr-tester
ct=$?

exit $cr || $cl || $cc || $ct
