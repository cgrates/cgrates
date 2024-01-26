#! /usr/bin/env sh
echo "Building CGRateS ..."

GIT_COMMIT="HEAD"

GIT_COMMIT_DATE="$(git log -n1 --format=format:%cI "${GIT_COMMIT}")"
GIT_COMMIT_HASH="$(git log -n1 --format=format:%H "${GIT_COMMIT}")"

GIT_TAG_LOG="$(git tag -l --points-at "${GIT_COMMIT}")"

if [ -n "${GIT_TAG_LOG}" ]; then
    GIT_COMMIT_DATE=""
    GIT_COMMIT_HASH=""
fi

go install -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                     -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
           github.com/cgrates/cgrates/cmd/cgr-engine
cr=$?
go install -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                     -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
           github.com/cgrates/cgrates/cmd/cgr-loader
cl=$?
go install -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                     -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
           github.com/cgrates/cgrates/cmd/cgr-console
cc=$?
go install -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                     -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
           github.com/cgrates/cgrates/cmd/cgr-migrator
cm=$?
go install -ldflags "-X 'github.com/cgrates/cgrates/utils.GitCommitDate=$GIT_COMMIT_DATE' \
                     -X 'github.com/cgrates/cgrates/utils.GitCommitHash=$GIT_COMMIT_HASH'" \
           github.com/cgrates/cgrates/cmd/cgr-tester
ct=$?

# shellcheck disable=SC2317
exit $cr || $cl || $cc || $cm || $ct
