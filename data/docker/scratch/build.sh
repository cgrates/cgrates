#! /usr/bin/env sh
echo "Static building CGRateS..."

GIT_COMMIT="HEAD"

GIT_COMMIT_DATE="$(git log -n1 --format=format:%cI "${GIT_COMMIT}")"
GIT_COMMIT_HASH="$(git log -n1 --format=format:%H "${GIT_COMMIT}")"

GIT_TAG_LOG="$(git tag -l --points-at "${GIT_COMMIT}")"

if [ -n "${GIT_TAG_LOG}" ]; then
    GIT_COMMIT_DATE=""
    GIT_COMMIT_HASH=""
fi

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-engine -a \
                                               -ldflags '-extldflags "-f no-PIC -static"' \
                                               -tags 'osusergo netgo static_build' \
                                               -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                                                         -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
                                               github.com/cgrates/cgrates/cmd/cgr-engine
cr=$?
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-loader -a \
                                               -ldflags '-extldflags "-f no-PIC -static"' \
                                               -tags 'osusergo netgo static_build' \
                                               -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                                                         -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
                                               github.com/cgrates/cgrates/cmd/cgr-loader
cl=$?
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-console -a \
                                               -ldflags '-extldflags "-f no-PIC -static"' \
                                               -tags 'osusergo netgo static_build' \
                                               -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                                                         -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
                                               github.com/cgrates/cgrates/cmd/cgr-console
cc=$?
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-migrator -a \
                                               -ldflags '-extldflags "-f no-PIC -static"' \
                                               -tags 'osusergo netgo static_build' \
                                               -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                                                         -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
                                               github.com/cgrates/cgrates/cmd/cgr-migrator
cm=$?
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cgr-tester -a \
                                               -ldflags '-extldflags "-f no-PIC -static"' \
                                               -tags 'osusergo netgo static_build' \
                                               -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                                                         -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
                                               github.com/cgrates/cgrates/cmd/cgr-tester
ct=$?

# shellcheck disable=SC2317
exit $cr || $cl || $cc || $cm || $ct
