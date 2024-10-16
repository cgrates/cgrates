FROM debian:bullseye-slim

RUN groupadd -r mongodb && useradd -r -g mongodb mongodb
RUN groupadd -r redis && useradd -r -g redis redis 
RUN groupadd -r kafka && useradd -r -g kafka kafka 
RUN groupadd -r nats && useradd -r -g nats nats

RUN mkdir -p /usr/share/cgrates/conf /usr/share/man/man1 /var/spool/cgrates/ers/in /var/spool/cgrates/ers/out /var/spool/cgrates/cdre/csv /var/spool/cgrates/cdre/fwv /var/spool/cgrates/tpe /var/spool/cgrates/failed_posts /var/spool/cgrates/analyzers /run /data/configdb /data/db /kafka /logs /opt/nats

RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

RUN apt-get update

RUN apt-get install -y apt-utils wget gnupg gnupg2 apt-transport-https curl redis-server git gcc build-essential rsyslog procps gosu "mariadb-server" mariadb-backup socat default-jdk neovim net-tools

RUN wget https://github.com/nats-io/nats-server/releases/download/v2.9.17/nats-server-v2.9.17-linux-amd64.tar.gz -O /tmp/nats-server-v2.9.17-linux-amd64.tar.gz && tar -xzvf /tmp/nats-server-v2.9.17-linux-amd64.tar.gz -C /opt/nats --strip-components=1

RUN ln -s /opt/nats/nats-server /usr/local/bin/nats-server

RUN rm /tmp/nats-server-v2.9.17-linux-amd64.tar.gz

RUN curl -fsSL https://pgp.mongodb.com/server-7.0.asc | gpg -o /usr/share/keyrings/mongodb-server-7.0.gpg  --dearmor
RUN echo "deb [ signed-by=/usr/share/keyrings/mongodb-server-7.0.gpg ] https://repo.mongodb.org/apt/debian bullseye/mongodb-org/7.0 main" | tee /etc/apt/sources.list.d/mongodb-org-7.0.list

RUN curl -1sLf "https://keys.openpgp.org/vks/v1/by-fingerprint/0A9AF2115F4687BD29803A206B73A36E6026DFCA" | gpg --dearmor | tee /usr/share/keyrings/com.rabbitmq.team.gpg > /dev/null
RUN curl -1sLf "https://github.com/rabbitmq/signing-keys/releases/download/3.0/cloudsmith.rabbitmq-erlang.E495BB49CC4BBE5B.key" | gpg --dearmor | tee /usr/share/keyrings/rabbitmq.E495BB49CC4BBE5B.gpg > /dev/null
RUN curl -1sLf "https://github.com/rabbitmq/signing-keys/releases/download/3.0/cloudsmith.rabbitmq-server.9F4587F226208342.key" | gpg --dearmor | tee /usr/share/keyrings/rabbitmq.9F4587F226208342.gpg > /dev/null
RUN tee /etc/apt/sources.list.d/rabbitmq.list <<EOF
## Provides modern Erlang/OTP releases
##
deb [signed-by=/usr/share/keyrings/rabbitmq.E495BB49CC4BBE5B.gpg] https://ppa1.novemberain.com/rabbitmq/rabbitmq-erlang/deb/debian bullseye main
deb-src [signed-by=/usr/share/keyrings/rabbitmq.E495BB49CC4BBE5B.gpg] https://ppa1.novemberain.com/rabbitmq/rabbitmq-erlang/deb/debian bullseye main

# another mirror for redundancy
deb [signed-by=/usr/share/keyrings/rabbitmq.E495BB49CC4BBE5B.gpg] https://ppa2.novemberain.com/rabbitmq/rabbitmq-erlang/deb/debian bullseye main
deb-src [signed-by=/usr/share/keyrings/rabbitmq.E495BB49CC4BBE5B.gpg] https://ppa2.novemberain.com/rabbitmq/rabbitmq-erlang/deb/debian bullseye main

## Provides RabbitMQ
##
deb [signed-by=/usr/share/keyrings/rabbitmq.9F4587F226208342.gpg] https://ppa1.novemberain.com/rabbitmq/rabbitmq-server/deb/debian bullseye main
deb-src [signed-by=/usr/share/keyrings/rabbitmq.9F4587F226208342.gpg] https://ppa1.novemberain.com/rabbitmq/rabbitmq-server/deb/debian bullseye main

# another mirror for redundancy
deb [signed-by=/usr/share/keyrings/rabbitmq.9F4587F226208342.gpg] https://ppa2.novemberain.com/rabbitmq/rabbitmq-server/deb/debian bullseye main
deb-src [signed-by=/usr/share/keyrings/rabbitmq.9F4587F226208342.gpg] https://ppa2.novemberain.com/rabbitmq/rabbitmq-server/deb/debian bullseye main
EOF

RUN apt-get update
RUN apt-get install -y erlang-base erlang-asn1 erlang-crypto erlang-eldap erlang-ftp erlang-inets erlang-mnesia erlang-os-mon erlang-parsetools erlang-public-key erlang-runtime-tools erlang-snmp erlang-ssl erlang-syntax-tools erlang-tftp erlang-tools erlang-xmerl
RUN apt-get install -y  rabbitmq-server mongodb-org postgresql --fix-missing

WORKDIR /kafka

RUN wget "https://dlcdn.apache.org/kafka/3.5.0/kafka_2.13-3.5.0.tgz" 
RUN tar -xvzf ./kafka_2.13-3.5.0.tgz --strip 1
RUN rm kafka_2.13-3.5.0.tgz
COPY ./conf/server.properties /kafka/config/server.properties
WORKDIR /

RUN set -ex; \
	rm -rf /var/lib/apt/lists/*; \
	rm -rf /var/lib/mysql; \
	mkdir -p /var/lib/mysql /var/run/mysqld; \
	chown -R mysql:mysql /var/lib/mysql /var/run/mysqld; \
	chmod 777 /var/run/mysqld; \
	find /etc/mysql/ -name '*.cnf' -print0 \
		| xargs -0 grep -lZE '^(bind-address|log|user\s)' \
		| xargs -rt -0 sed -Ei 's/^(bind-address|log|user\s)/#&/'; \
	echo '[mysqld]\nskip-host-cache\nskip-name-resolve' > /etc/mysql/conf.d/docker.cnf

RUN apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN touch /logs/mariadb.log /logs/mariadb_script.log /logs/rabbitmq.log 
RUN chmod 777  /logs/mariadb.log /logs/mariadb_script.log /logs/rabbitmq.log

RUN wget -O go.tgz "https://storage.googleapis.com/golang/go1.23.2.linux-amd64.tar.gz" --progress=dot:giga
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
