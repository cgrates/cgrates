#! /usr/bin/env sh
echo "Static building CGRateS..."

GIT_LAST_LOG=$(git log -1 | tr -d "'")

GIT_TAG_LOG=$(git tag -l --points-at HEAD)

if [ ! -z "$GIT_TAG_LOG" ]
then
    GIT_LAST_LOG=""
fi

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-engine -a -ldflags '-extldflags "-f no-PIC -static"' -tags 'osusergo netgo static_build' -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-engine
cr=$?
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-loader -a -ldflags '-extldflags "-f no-PIC -static"' -tags 'osusergo netgo static_build' -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-loader
cl=$?
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-console -a -ldflags '-extldflags "-f no-PIC -static"' -tags 'osusergo netgo static_build' -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-console
cc=$?
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-migrator -a -ldflags '-extldflags "-f no-PIC -static"' -tags 'osusergo netgo static_build' -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-migrator
cm=$?
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-tester -a -ldflags '-extldflags "-f no-PIC -static"' -tags 'osusergo netgo static_build' -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-tester
ct=$?

exit $cr || $cl || $cc || $cm || $ct
