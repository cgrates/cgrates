#! /usr/bin/env sh

go install github.com/cgrates/cgrates/cmd/cgr-rater
go install github.com/cgrates/cgrates/cmd/cgr-loader
go install github.com/cgrates/cgrates/cmd/cgr-console

go test github.com/cgrates/cgrates/timespans
go test github.com/cgrates/cgrates/sessionmanager
go test github.com/cgrates/cgrates/balancer
go test github.com/cgrates/cgrates/cmd/cgr-rater
