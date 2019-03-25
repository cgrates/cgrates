Docker
=======

From the project root:

``` bash
# build the image
docker build -t cgrates data/docker/prod
# create the container
docker run --rm -itv `pwd`:/root/code --name cgr cgrates
# If running on local machine, do this to expose agent for /pprof:
docker run --rm -itv `pwd`:/root/code -p 127.0.0.1:2080:2080/tcp --name cgr cgrates
```
