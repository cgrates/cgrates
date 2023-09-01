#!/bin/bash
set -ev

ansible-playbook /integration_tests.yaml -i localhost, -c local -e "ansible_user=root clone_repository=true" -vvv

# # Create symbolic links
ln -s "/go/src/github.com/cgrates/cgrates/data" "/usr/share/cgrates"
ln -s "/go/bin/cgr-engine" "/usr/bin/cgr-engine"
ln -s "/go/bin/cgr-loader" "/usr/bin/cgr-loader"
ln -s "/go/bin/cgr-migrator" "/usr/bin/cgr-migrator"
ln -s "/go/bin/cgr-console" "/usr/bin/cgr-console"
ln -s "/go/bin/cgr-tester" "/usr/bin/cgr-tester"

# start basic subsystems
# export KAFKA_HEAP_OPTS="-Xmx100M -Xms100M"
# /kafka/bin/zookeeper-server-start.sh -daemon /kafka/config/zookeeper.properties
# /kafka/bin/kafka-server-start.sh -daemon /kafka/config/server.properties

rsyslogd -f /etc/rsyslogd.conf 
version=$(ls /var/lib/postgresql)
pg_ctlcluster $version main start &
mongod --bind_ip 127.0.0.1  --logpath /logs/mongodb.log &
redis-server /etc/redis/redis.conf &
MYSQL_ROOT_PASSWORD="CGRateS.org" /go/src/github.com/cgrates/cgrates/data/docker/integration/scripts/mariadb-ep.sh mysqld
# rabbitmq-server > /logs/rabbitmq.log  2>&1 &


# START_TIMEOUT=600

# start_timeout_exceeded=false
# count=0
# step=10
# while netstat -lnt | awk '$4 ~ /:9092$/ {exit 1}'; do
#     echo "waiting for kafka to be ready"
#     sleep $step;
#     count=$((count + step))
#     if [ $count -gt $START_TIMEOUT ]; then
#         start_timeout_exceeded=true
#         break
#     fi
# done

# if $start_timeout_exceeded; then
#     echo "Not able to auto-create topic (waited for $START_TIMEOUT sec)"
#     exit 1
# fi

# /kafka/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic cgrates
# /kafka/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic cgrates_cdrs



# gosu postgres psql  -c "CREATE USER cgrates password 'CGRateS.org';"  > /dev/null 2>&1
# gosu postgres createdb -e -O cgrates cgrates > /dev/null 2>&1



# PGPASSWORD="CGRateS.org" psql -U "cgrates" -h "localhost" -d cgrates -f /cgrates/data/docker/integration/scripts/postgres/create_cdrs_tables.sql >/dev/null 2>&1
# PGPASSWORD="CGRateS.org" psql -U "cgrates" -h "localhost" -d cgrates -f /cgrates/data/docker/integration/scripts/postgres/create_tariffplan_tables.sql >/dev/null 2>&1


# mongo --quiet /cgrates/data/docker/integration/scripts/create_user.js >/dev/null 2>&1
echo "Starting..."
sleep 60  # Pause for 1 min
echo "Resuming..."
mysql -u root -pCGRateS.org -h localhost < /go/src/github.com/cgrates/cgrates/data/docker/integration/scripts/mysql/create_db_with_users.sql 
mysql -u root -pCGRateS.org -h localhost -D cgrates < /go/src/github.com/cgrates/cgrates/data/docker/integration/scripts/mysql/create_cdrs_tables.sql
mysql -u root -pCGRateS.org -h localhost -D cgrates < /go/src/github.com/cgrates/cgrates/data/docker/integration/scripts/mysql/create_tariffplan_tables.sql

cp -r /go/src/github.com/cgrates/cgrates/data/. /usr/share/cgrates

# Set versions
# cgr-migrator -exec=*set_versions -config_path=/usr/share/cgrates/conf/samples/tutmysql