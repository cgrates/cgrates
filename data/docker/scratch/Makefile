
generate:
	./build.sh	
	cp ../../conf/cgrates/cgrates.json ./cgrates.json
build:
	docker build -t cgrates .
run: build
	docker run --rm -itv `pwd`:/root/code -p 2080:2080 -p 2012:2012 --name cgr cgrates
.PHONY: generate,build,run
