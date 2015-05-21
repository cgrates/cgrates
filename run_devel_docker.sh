#!/usr/bin/env sh

docker run --rm -p 3306:3306 -p 6379:6379 -p 2012:2012 -itv /home/rif/Documents/prog/go/src/github.com/cgrates/cgrates:/root/code/src/github.com/cgrates/cgrates --name cgr cgrates
