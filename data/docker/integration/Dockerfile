FROM  debian:buster-slim

RUN groupadd -r mongodb && useradd -r -g mongodb mongodb
RUN groupadd -r redis && useradd -r -g redis redis 
RUN groupadd -r kafka && useradd -r -g kafka kafka 

RUN mkdir -p /usr/share/man/man1 /var/spool/cgrates/ers/in /var/spool/cgrates/ers/out /var/spool/cgrates/cdre/csv /var/spool/cgrates/cdre/fwv /var/spool/cgrates/tpe /var/spool/cgrates/failed_posts /var/spool/cgrates/analyzers /run /data/configdb /data/db /kafka /logs

RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# Install necessary libs
RUN apt-get update 
RUN apt-get install -y apt-utils wget gnupg gnupg2 apt-transport-https curl redis-server git build-essential rsyslog procps gosu "mariadb-server" mariadb-backup socat default-jdk-headless neovim

RUN wget -qO - https://www.mongodb.org/static/pgp/server-4.4.asc | apt-key add -
RUN echo "deb http://repo.mongodb.org/apt/debian buster/mongodb-org/4.4 main" | tee /etc/apt/sources.list.d/mongodb-org-4.4.list

RUN wget -qO - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add -
RUN echo "deb http://apt.postgresql.org/pub/repos/apt buster-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list

RUN wget -qO - https://github.com/rabbitmq/signing-keys/releases/download/2.0/rabbitmq-release-signing-key.asc | apt-key add -
RUN printf "deb https://dl.bintray.com/rabbitmq-erlang/debian buster erlang\ndeb https://dl.bintray.com/rabbitmq/debian buster main" | tee /etc/apt/sources.list.d/bintray.rabbitmq.list

RUN apt-get update
RUN apt-get install -y mongodb-org postgresql rabbitmq-server


WORKDIR /kafka

RUN wget "https://archive.apache.org/dist/kafka/2.1.1/kafka_2.11-2.1.1.tgz" 
RUN tar -xvzf ./kafka_2.11-2.1.1.tgz --strip 1
RUN rm kafka_2.11-2.1.1.tgz
COPY ./conf/server.properties /kafka/config/server.properties
WORKDIR /


RUN set -ex; \
	rm -rf /var/lib/apt/lists/*; \
# purge and re-create /var/lib/mysql with appropriate ownership
	rm -rf /var/lib/mysql; \
	mkdir -p /var/lib/mysql /var/run/mysqld; \
	chown -R mysql:mysql /var/lib/mysql /var/run/mysqld; \
# ensure that /var/run/mysqld (used for socket and lock files) is writable regardless of the UID our mysqld instance ends up having at runtime
	chmod 777 /var/run/mysqld; \
# comment out a few problematic configuration values
	find /etc/mysql/ -name '*.cnf' -print0 \
		| xargs -0 grep -lZE '^(bind-address|log|user\s)' \
		| xargs -rt -0 sed -Ei 's/^(bind-address|log|user\s)/#&/'; \
# don't reverse lookup hostnames, they are usually another container
	echo '[mysqld]\nskip-host-cache\nskip-name-resolve' > /etc/mysql/conf.d/docker.cnf


RUN wget -O go.tgz "https://storage.googleapis.com/golang/go1.15.7.linux-amd64.tar.gz" --progress=dot:giga
RUN tar -C /usr/local -xzf go.tgz
RUN rm go.tgz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
RUN	go version


RUN touch /logs/mariadb.log /logs/mariadb_script.log /logs/rabbitmq.log 
RUN chmod 777  /logs/mariadb.log /logs/mariadb_script.log /logs/rabbitmq.log 


COPY ./scripts /scripts
COPY ./scripts/service /usr/local/bin/service

COPY ./conf/rsyslogd.conf /etc/rsyslogd.conf
COPY ./conf/rsyslog.d /etc/rsyslog.d
COPY ./conf/redis.conf /etc/redis/redis.conf

COPY ./docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]