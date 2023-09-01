FROM  ubuntu:jammy

RUN groupadd -r mongodb && useradd -r -g mongodb mongodb
RUN groupadd -r redis && useradd -r -g redis redis 
RUN groupadd -r kafka && useradd -r -g kafka kafka 

RUN mkdir -p /usr/share/cgrates/conf /usr/share/man/man1 /var/spool/cgrates/ers/in /var/spool/cgrates/ers/out /var/spool/cgrates/cdre/csv /var/spool/cgrates/cdre/fwv /var/spool/cgrates/tpe /var/spool/cgrates/failed_posts /var/spool/cgrates/analyzers /run /data/configdb /data/db /kafka /logs


ENV DEBIAN_FRONTEND=noninteractive


RUN apt-get update && apt-get install -y ansible

# Install Erlang packages
RUN apt-get install -y erlang-base erlang-asn1 erlang-crypto erlang-eldap erlang-ftp erlang-inets erlang-mnesia erlang-os-mon erlang-parsetools erlang-public-key erlang-runtime-tools erlang-snmp erlang-ssl erlang-syntax-tools erlang-tftp erlang-tools erlang-xmerl

RUN apt-get install -y apt-utils wget gnupg gnupg2 apt-transport-https curl redis-server git build-essential rsyslog procps gosu "mariadb-server" mariadb-backup socat default-jdk-headless neovim net-tools

RUN curl -fsSL https://pgp.mongodb.com/server-7.0.asc | gpg -o /usr/share/keyrings/mongodb-server-7.0.gpg  --dearmor
RUN echo "deb [ arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/7.0 multiverse" | tee /etc/apt/sources.list.d/mongodb-org-7.0.list

RUN sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
RUN wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc |  apt-key add -

RUN apt-get update
RUN apt-get install -y mongodb-org postgresql
# ADD THIS AFTER postgresql rabbitmq-server

WORKDIR /kafka

RUN wget "https://archive.apache.org/dist/kafka/2.1.0/kafka_2.11-2.1.0.tgz" 
RUN tar -xvzf ./kafka_2.11-2.1.0.tgz --strip 1
RUN rm kafka_2.11-2.1.0.tgz
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

# Clean up
RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN touch /logs/mariadb.log /logs/mariadb_script.log /logs/rabbitmq.log 
RUN chmod 777  /logs/mariadb.log /logs/mariadb_script.log /logs/rabbitmq.log 

COPY main.yaml /integration_tests.yaml

COPY roles /roles

RUN wget -O go.tgz "https://storage.googleapis.com/golang/go1.21.0.linux-amd64.tar.gz" --progress=dot:giga
RUN tar -C /usr/local -xzf go.tgz
RUN rm go.tgz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
RUN	go version

COPY ./scripts /scripts
COPY ./scripts/service /usr/local/bin/service

COPY ./conf/rsyslogd.conf /etc/rsyslogd.conf
COPY ./conf/rsyslog.d /etc/rsyslog.d
COPY ./conf/redis.conf /etc/redis/redis.conf

COPY ./docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]