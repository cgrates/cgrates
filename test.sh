#! /usr/bin/env sh

go test github.com/cgrates/cgrates/rater
ts=$?
go test github.com/cgrates/cgrates/sessionmanager
sm=$?
go test github.com/cgrates/cgrates/cmd/cgr-rater
cr=$?
go test github.com/cgrates/cgrates/inotify
it=$?
go test github.com/cgrates/cgrates/mediator
md=$?

exit $ts && $sm && $bl && $cr && $it && $md
