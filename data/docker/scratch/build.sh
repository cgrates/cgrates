#! /usr/bin/env sh
echo "Building CGRateS ..."

GIT_LAST_LOG=$(git log -1 | tr -d "'")

GIT_TAG_LOG=$(git tag -l --points-at HEAD)

if [ ! -z "$GIT_TAG_LOG" ]
then
    GIT_LAST_LOG=""
fi

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-engine -a -ldflags '-extldflags "-f no-PIC -static"' -tags 'osusergo netgo static_build' -ldflags "-X 'github.com/cgrates/cgrates/utils.GitLastLog=$GIT_LAST_LOG'" github.com/cgrates/cgrates/cmd/cgr-engine
cr=$?

exit $cr 
