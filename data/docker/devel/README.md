Docker
=======

From the project root:

``` bash
# build the image
docker build -t cgrates data/docker/devel
# create the container
docker run --rm -itv `pwd`:/root/code/src/github.com/cgrates/cgrates --name cgr cgrates
```
