#! /usr/bin/env sh

go install github.com/cgrates/cgrates/cmd/cgr-rater
go install github.com/cgrates/cgrates/cmd/cgr-loader
go install github.com/cgrates/cgrates/cmd/cgr-console

go test github.com/cgrates/cgrates/timespans
ts=$?
go test github.com/cgrates/cgrates/sessionmanager
sm=$?
go test github.com/cgrates/cgrates/balancer
bl=$?
go test github.com/cgrates/cgrates/cmd/cgr-rater
cr=$?

exit $ts || $sm || $bl || $cr


