FROM debian:latest
MAINTAINER Radu Fericean, rif@cgrates.org

# set mysql password
RUN echo 'mysql-server mysql-server/root_password password CGRateS.org' | debconf-set-selections && echo 'mysql-server mysql-server/root_password_again password CGRateS.org' | debconf-set-selections

# install dependencies
RUN apt-get -y update && apt-get -y install git redis-server mysql-server python-pycurl python-mysqldb sudo wget vim zsh tmux rsyslog ngrep curl

# add cgrates user
RUN useradd -c CGRateS -d /var/run/cgrates -s /bin/false -r cgrates

# install golang
RUN wget -qO- https://storage.googleapis.com/golang/go1.6.3.linux-amd64.tar.gz | tar xzf - -C /root/

#install glide
RUN GOROOT=/root/go GOPATH=/root/code /root/go/bin/go get github.com/Masterminds/glide

#install oh-my-zsh
RUN TERM=xterm sh -c "$(wget https://raw.github.com/robbyrussell/oh-my-zsh/master/tools/install.sh -O -)"; exit 0

# change shell for tmux
RUN chsh -s /usr/bin/zsh

# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# create a link to data dir
RUN ln -s /root/code/src/github.com/cgrates/cgrates/data /usr/share/cgrates

# create link to cgrates dir
RUN ln -s /root/code/src/github.com/cgrates/cgrates /root/cgr

# setup mysql for CGRateS
RUN cd /usr/share/cgrates/storage/mysql && ./setup_cgr_db.sh root CGRateS.org


#env vars
RUN export GOROOT=/root/go; export GOPATH=/root/code; export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

# build and install cgrates
RUN cd /root/cgr
RUN glide install

# create cgr-engine and cgr-loader link
RUN ln -s /root/code/bin/cgr-engine /usr/bin/
RUN ln -s /root/code/bin/cgr-loader /usr/bin/


RUN cd /root/cgr
RUN echo 'export GOROOT=/root/go; export GOPATH=/root/code; export PATH=$GOROOT/bin:$GOPATH/bin:$PATH'>>/root/.zshrc

# set start command
RUN CMD /root/code/src/github.com/cgrates/cgrates/data/docker/osips_training/start.sh
