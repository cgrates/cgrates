#! /usr/bin/env sh

go test -i github.com/cgrates/cgrates/rater
go test -i github.com/cgrates/cgrates/sessionmanager
go test -i github.com/cgrates/cgrates/config
go test -i github.com/cgrates/cgrates/cmd/cgr-rater
go test -i github.com/cgrates/cgrates/inotify
go test -i github.com/cgrates/cgrates/mediator
go test -i github.com/cgrates/fsock


go test github.com/cgrates/cgrates/rater
ts=$?
go test github.com/cgrates/cgrates/sessionmanager
sm=$?
go test github.com/cgrates/cgrates/config
cfg=$?
go test github.com/cgrates/cgrates/cmd/cgr-rater
cr=$?
go test github.com/cgrates/cgrates/inotify
it=$?
go test github.com/cgrates/cgrates/mediator
md=$?
go test github.com/cgrates/fsock
fs=$?

exit $ts && $sm && $cfg && $bl && $cr && $it && $md && $fs
