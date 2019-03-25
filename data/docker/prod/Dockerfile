FROM debian:latest
LABEL maintainer="Radu Fericean, rif@cgrates.org, Gavin Henry, ghenry@surevoip.co.uk"
RUN apt-get update && apt-get install -y gnupg2 wget apt-utils ngrep vim

# set mysql password
RUN echo 'mysql-server mysql-server/root_password password CGRateS.org' | debconf-set-selections && echo 'mysql-server mysql-server/root_password_again password CGRateS.org' | debconf-set-selections

# add freeswitch gpg key
RUN wget -O - https://files.freeswitch.org/repo/deb/freeswitch-1.8/fsstretch-archive-keyring.asc | apt-key add -

# add freeswitch apt repo
RUN echo "deb http://files.freeswitch.org/repo/deb/freeswitch-1.8/ stretch main" > /etc/apt/sources.list.d/freeswitch.list

# install dependencies
RUN apt-get update && apt-get -y install redis-server mysql-server python-pycurl python-mysqldb postgresql postgresql-client sudo wget git freeswitch-meta-vanilla

####
# Re-enable this once the CGRateS repo is live - GH.
#
# add cgrates apt-key
#RUN wget -qO- http://apt.itsyscom.com/conf/cgrates.gpg.key | apt-key add -

# add cgrates repo
#RUN cd /etc/apt/sources.list.d/; wget -q http://apt.itsyscom.com/conf/cgrates.apt.list

# install cgrates
#RUN apt-get update && apt-get -y install cgrates

# CGRateS 
RUN wget -O /tmp/cgrates.deb http://www.cgrates.org/tmp_pkg/cgrates_0.9.1~rc8_amd64.deb
RUN apt install /tmp/cgrates.deb
RUN rm /tmp/cgrates.deb

# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# set start command
CMD /root/code/data/docker/prod/start.sh

