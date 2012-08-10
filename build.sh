#! /usr/bin/env sh

go install github.com/cgrates/cgrates/cmd/cgr-rater
cr=$?
go install github.com/cgrates/cgrates/cmd/cgr-loader
cl=$?
go install github.com/cgrates/cgrates/cmd/cgr-console
cc=$?

exit $cr || $cl || $cc


