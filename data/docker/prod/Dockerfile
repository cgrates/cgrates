FROM debian:latest
MAINTAINER Radu Fericean, rif@cgrates.org
RUN apt-get -y update

# install dependencies
RUN apt-get -y install git redis-server mysql-server python-mysqldb wget

# add freeswitch gpg key
RUN gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF
      
# add freeswitch apt key
RUN gpg -a --export D76EDC7725E010CF | apt-key add -

# add freeswitch apt repo
RUN cd /etc/apt/sources.list.d/; wget http://apt.itsyscom.com/conf/freeswitch.apt.list
      
# install freeswitch
RUN apt-get update; apt-get -y install freeswitch-meta-vanilla freeswitch-mod-json-cdr

# add cgrates apt-key
RUN wget -O - http://apt.itsyscom.com/conf/cgrates.gpg.key | apt-key add -

# add cgrates repo
RUN cd /etc/apt/sources.list.d/; wget http://apt.itsyscom.com/conf/cgrates.apt.list

# install cgrates
RUN apt-get update; apt-get -y install cgrates

# copy .my.cnf  (do not enable it)
ADD my.cnf /my.cnf

# add startup script
ADD start.sh /root/start.sh

# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# set start command
CMD /root/start.sh

