FROM debian:latest
MAINTAINER Radu Fericean, rif@cgrates.org

# set mysql password
RUN echo 'mysql-server mysql-server/root_password password CGRateS.org' | debconf-set-selections && echo 'mysql-server mysql-server/root_password_again password CGRateS.org' | debconf-set-selections

# add freeswitch gpg key
RUN gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF && gpg -a --export D76EDC7725E010CF | apt-key add -

# add freeswitch apt repo
RUN echo 'deb http://files.freeswitch.org/repo/deb/debian/ jessie main' > /etc/apt/sources.list.d/freeswitch.list
    
# install dependencies
RUN apt-get -y update && apt-get -y install git bzr mercurial redis-server mysql-server python-pycurl python-mysqldb postgresql postgresql-client sudo wget freeswitch-meta-vanilla

# add cgrates user
RUN useradd -c CGRateS -d /var/run/cgrates -s /bin/false -r cgrates      

# install golang
RUN wget -qO- https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar xzf - -C /root/

# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# set start command
CMD /root/code/src/github.com/cgrates/cgrates/data/docker/devel/start.sh
