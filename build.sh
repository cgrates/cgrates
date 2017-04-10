#! /usr/bin/env sh
echo "Building CGRateS ..."

go install github.com/cgrates/cgrates/cmd/cgr-engine
go install github.com/cgrates/cgrates/cmd/cgr-tester
go install github.com/cgrates/cgrates/cmd/cgr-console
go install github.com/cgrates/cgrates/cmd/cgr-loader


GIT_LAST_LOG=$(git log -1)
go install -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-engine
cr=$?
go install -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-loader
cl=$?
go install -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-console
cc=$?
go install -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-tester
ct=$?

exit $cr || $cl || $cc || $ct
