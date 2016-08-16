FROM debian:latest
MAINTAINER Radu Fericean, rif@cgrates.org

# set mysql password
RUN echo 'mysql-server mysql-server/root_password password CGRateS.org' | debconf-set-selections && echo 'mysql-server mysql-server/root_password_again password CGRateS.org' | debconf-set-selections

# install dependencies
RUN apt-get -y update && apt-get -y install git redis-server mysql-server python-pycurl python-mysqldb sudo wget vim zsh tmux rsyslog ngrep curl

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

# get cgrates from github
RUN mkdir -p /root/code/src/github.com/cgrates; cd /root/code/src/github.com/cgrates; git clone https://github.com/cgrates/cgrates.git

# create link to cgrates dir
RUN ln -s /root/code/src/github.com/cgrates/cgrates /root/cgr

# get deps
RUN cd /root/cgr; GOROOT=/root/go GOPATH=/root/code PATH=$GOROOT/bin:$GOPATH/bin:$PATH glide install

# build cgrates
RUN cd /root/cgr; GOROOT=/root/go GOPATH=/root/code PATH=$GOROOT/bin:$GOPATH/bin:$PATH ./build.sh

# create links
RUN ln -s /root/code/bin/cgr-engine /usr/bin/
RUN ln -s /root/code/bin/cgr-loader /usr/bin/
RUN ln -s /root/code/src/github.com/cgrates/cgrates/data /usr/share/cgrates

# prepare zshrc
RUN echo 'export GOROOT=/root/go; export GOPATH=/root/code; export PATH=$GOROOT/bin:$GOPATH/bin:$PATH'>>/root/.zshrc

# set start command
CMD /root/code/src/github.com/cgrates/cgrates/data/docker/osips_training/start.sh
