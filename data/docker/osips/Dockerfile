FROM debian:wheezy
MAINTAINER Radu Fericean, rif@cgrates.org
RUN apt-get -y update

# set mysql password
RUN echo 'mysql-server mysql-server/root_password password CGRateS.org' | debconf-set-selections && echo 'mysql-server mysql-server/root_password_again password CGRateS.org' | debconf-set-selections
    
# install dependencies
RUN apt-get update && apt-get -y install redis-server mysql-server git sudo wget libcurl3-gnutls libgssapi-krb5-2 libldap-2.4-2 librtmp0 libssh2-1 ca-certificates libsasl2-2 libkrb5support0 libkrb5-3 libkeyutils1 libk5crypto3 openssl libssl1.0.0 libjson0 wget adduser libcurl3-gnutls

# add cgrates user
RUN useradd -c CGRateS -d /var/run/cgrates -s /bin/false -r cgrates
    
# install cgrates
RUN URL='http://www.cgrates.org/tmp_pkg/cgrates_0.9.1~rc6_amd64.deb'; FILE=`mktemp`; wget "$URL" -qO $FILE && dpkg -i $FILE; rm $FILE

#install opensips
RUN URL='http://cgrates.org/pkg/opensips/opensips_2.1.0-1_amd64.deb'; FILE=`mktemp`; wget "$URL" -qO $FILE && dpkg -i $FILE; rm $FILE
RUN URL='http://cgrates.org/pkg/opensips/opensips-json-module_2.1.0-1_amd64.deb'; FILE=`mktemp`; wget "$URL" -qO $FILE && dpkg -i $FILE; rm $FILE

# init mysql
cd /usr/share/cgrates/storage/mysql && ./setup_cgr_db.sh root CGRateS.org
    
# add etc files
COPY /opensips/etc cgrates/etc /etc
    
# cleanup
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# set start command
CMD /root/code/data/docker/prod/start.sh

