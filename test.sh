#! /usr/bin/env sh

go test github.com/cgrates/cgrates/timespans
ts=$?
go test github.com/cgrates/cgrates/sessionmanager
sm=$?
go test github.com/cgrates/cgrates/cmd/cgr-rater
cr=$?

exit $ts || $sm || $bl || $cr


