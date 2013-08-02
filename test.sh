#! /usr/bin/env sh

go test -i github.com/cgrates/cgrates/engine
go test -i github.com/cgrates/cgrates/sessionmanager
go test -i github.com/cgrates/cgrates/config
go test -i github.com/cgrates/cgrates/cmd/cgr-engine
go test -i github.com/cgrates/cgrates/mediator
go test -i github.com/cgrates/fsock
go test -i github.com/cgrates/cgrates/cdrs
go test -i github.com/cgrates/cgrates/utils
go test -i github.com/cgrates/cgrates/history

go test github.com/cgrates/cgrates/engine
en=$?
go test github.com/cgrates/cgrates/sessionmanager
sm=$?
go test github.com/cgrates/cgrates/config
cfg=$?
go test github.com/cgrates/cgrates/cmd/cgr-engine
cr=$?
go test github.com/cgrates/cgrates/mediator
md=$?
go test github.com/cgrates/cgrates/cdrs
cdr=$?
go test github.com/cgrates/cgrates/utils
ut=$?
go test github.com/cgrates/fsock
fs=$?
go test github.com/cgrates/cgrates/history
hs=$?

exit $en && $sm && $cfg && $bl && $cr && $md && $cdr && $fs && $ut && $hs
