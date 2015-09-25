FROM debian:latest
MAINTAINER Radu Fericean, rif@cgrates.org

# set mysql password
RUN echo 'mysql-server mysql-server/root_password password CGRateS.org' | debconf-set-selections && echo 'mysql-server mysql-server/root_password_again password CGRateS.org' | debconf-set-selections

# add freeswitch gpg key
RUN gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF && gpg -a --export D76EDC7725E010CF | apt-key add -

# add freeswitch apt repo
RUN echo 'deb http://files.freeswitch.org/repo/deb/debian/ jessie main' > /etc/apt/sources.list.d/freeswitch.list

# add cassandra gpg keys
RUN gpg --keyserver pgp.mit.edu --recv-keys F758CE318D77295D; gpg --export --armor F758CE318D77295D | apt-key add -;gpg --keyserver pgp.mit.edu --recv-keys 2B5C1B00;gpg --export --armor 2B5C1B00 |  apt-key add -;gpg --keyserver pgp.mit.edu --recv-keys 0353B12C;gpg --export --armor 0353B12C | apt-key add -

# add cassandra repo
RUN echo 'deb http://www.apache.org/dist/cassandra/debian 21x main' > /etc/apt/sources.list.d/cassandra.list

# install dependencies
RUN apt-get -y update && apt-get -y install git bzr mercurial redis-server mysql-server python-pycurl python-mysqldb postgresql postgresql-client cassandra sudo wget freeswitch-meta-vanilla vim zsh

# add cgrates user
RUN useradd -c CGRateS -d /var/run/cgrates -s /bin/false -r cgrates

# install golang
RUN wget -qO- https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz | tar xzf - -C /root/

#install glide
RUN GOROOT=/root/go GOPATH=/root/code /root/go/bin/go get github.com/Masterminds/glide

#install oh-my-zsh
RUN TERM=xterm sh -c "$(wget https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh -O -)"; exit 0

# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# set start command
CMD /root/code/src/github.com/cgrates/cgrates/data/docker/devel/start.sh
