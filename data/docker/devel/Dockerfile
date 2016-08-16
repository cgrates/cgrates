FROM debian:latest
MAINTAINER Radu Fericean, rif@cgrates.org

# set mysql password
RUN echo 'mysql-server mysql-server/root_password password CGRateS.org' | debconf-set-selections && echo 'mysql-server mysql-server/root_password_again password CGRateS.org' | debconf-set-selections

# add freeswitch gpg key
RUN gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF && gpg -a --export D76EDC7725E010CF | apt-key add -

# add freeswitch apt repo
RUN echo 'deb http://files.freeswitch.org/repo/deb/debian/ jessie main' > /etc/apt/sources.list.d/freeswitch.list

# add mongo repo keys
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv EA312927

# add mongo repo
RUN echo 'deb http://repo.mongodb.org/apt/debian wheezy/mongodb-org/3.2 main' | tee '/etc/apt/sources.list.d/mongodb-org-3.2.list'

# install dependencies
RUN apt-get -y update && apt-get -y install git redis-server mysql-server python-pycurl python-mysqldb postgresql postgresql-client sudo wget freeswitch-meta-vanilla vim zsh mongodb-org tmux rsyslog ngrep curl

# add mongo conf
COPY mongod.conf /etc/mongod.conf

# add cgrates user
RUN useradd -c CGRateS -d /var/run/cgrates -s /bin/false -r cgrates

# install golang
RUN wget -qO- https://storage.googleapis.com/golang/go1.7.linux-amd64.tar.gz | tar xzf - -C /root/

#install glide
RUN GOROOT=/root/go GOPATH=/root/code /root/go/bin/go get github.com/Masterminds/glide

#install oh-my-zsh
RUN TERM=xterm sh -c "$(wget https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh -O -)"; exit 0

# change shell for tmux
RUN chsh -s /usr/bin/zsh

# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# set start command
CMD /root/code/src/github.com/cgrates/cgrates/data/docker/devel/start.sh
