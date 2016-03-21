#! /usr/bin/env sh
./build.sh

go test -i github.com/cgrates/cgrates/apier/v1
go test -i github.com/cgrates/cgrates/apier/v2
go test -i github.com/cgrates/cgrates/engine
go test -i github.com/cgrates/cgrates/sessionmanager
go test -i github.com/cgrates/cgrates/config
go test -i github.com/cgrates/cgrates/cmd/cgr-engine
go test -i github.com/cgrates/cgrates/cache2go
go test -i github.com/cgrates/cgrates/cdrc
go test -i github.com/cgrates/cgrates/utils
go test -i github.com/cgrates/cgrates/history
go test -i github.com/cgrates/cgrates/cdre
go test -i github.com/cgrates/cgrates/agents
go test -i github.com/cgrates/cgrates/structmatcher

go test github.com/cgrates/cgrates/apier/v1
v1=$?
go test github.com/cgrates/cgrates/apier/v2
v2=$?
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
go test github.com/cgrates/cgrates/console
con=$?
go test github.com/cgrates/cgrates/cdrc
cdrcs=$?
go test github.com/cgrates/cgrates/utils
ut=$?
go test github.com/cgrates/cgrates/history
hs=$?
go test github.com/cgrates/cgrates/cache2go
c2g=$?
go test github.com/cgrates/cgrates/cdre
cdre=$?
go test github.com/cgrates/cgrates/agents
ag=$?
go test github.com/cgrates/cgrates/structmatcher
sc=$?


exit $v1 && $v2 && $en && $gt && $sm && $cfg && $bl && $cr && $con && $cdrc && $ut && $hs && $c2g && $cdre && $ag && $sc
