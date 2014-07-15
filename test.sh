#! /usr/bin/env sh

go test -i github.com/cgrates/cgrates/engine
go test -i github.com/cgrates/cgrates/sessionmanager
go test -i github.com/cgrates/cgrates/config
go test -i github.com/cgrates/cgrates/cmd/cgr-engine
go test -i github.com/cgrates/cgrates/mediator
go test -i github.com/cgrates/fsock
go test -i github.com/cgrates/cgrates/cache2go
go test -i github.com/cgrates/cgrates/cdrs
go test -i github.com/cgrates/cgrates/cdrc
go test -i github.com/cgrates/cgrates/utils
go test -i github.com/cgrates/cgrates/history
go test -i github.com/cgrates/cgrates/cdre

go test github.com/cgrates/cgrates/engine
en=$?
go test github.com/cgrates/cgrates/general_tests
gt=$?
go test github.com/cgrates/cgrates/sessionmanager
sm=$?
go test github.com/cgrates/cgrates/config
cfg=$?
go test github.com/cgrates/cgrates/cmd/cgr-engine
cr=$?
go test github.com/cgrates/cgrates/mediator
md=$?
go test github.com/cgrates/cgrates/cdrs
cdrs=$?
go test github.com/cgrates/cgrates/cdrc
cdrcs=$?
go test github.com/cgrates/cgrates/utils
ut=$?
go test github.com/cgrates/fsock
fs=$?
go test github.com/cgrates/cgrates/history
hs=$?
go test github.com/cgrates/cgrates/cache2go
c2g=$?
go test github.com/cgrates/cgrates/cdre
cdre=$?

exit $en && $gt && $sm && $cfg && $bl && $cr && $md && $cdrs && $cdrc && $fs && $ut && $hs && $c2g && $cdre
