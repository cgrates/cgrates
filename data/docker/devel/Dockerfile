FROM debian:latest
MAINTAINER Radu Fericean, rif@cgrates.org
RUN apt-get -y update

# install dependency
RUN apt-get -y install git bzr mercurial redis-server mysql-server python-pycurl python-mysqldb wget

# add cgrates user
RUN useradd -c CGRateS -d /var/run/cgrates -s /bin/false -r cgrates

# add freeswitch gpg key
RUN gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF
      
# add freeswitch apt key
RUN gpg -a --export D76EDC7725E010CF | apt-key add -

# add freeswitch apt repo
RUN cd /etc/apt/sources.list.d/; wget http://apt.itsyscom.com/conf/freeswitch.apt.list
      
# install freeswitch
RUN apt-get update; apt-get -y install freeswitch-meta-vanilla freeswitch-mod-json-cdr

# install golang
RUN wget -qO- https://storage.googleapis.com/golang/go1.3.linux-amd64.tar.gz | tar xzf - -C /root/

# add startup script
ADD start.sh /root/start.sh

# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# set start command
CMD /root/start.sh
