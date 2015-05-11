Docker
=======

From the project root:

``` bash
# build the image
docker build -t osips data/docker/ospis
# create the container
docker run --rm -itv `pwd`:/root/code --name cgr osips
```
