#!/bin/bash
set -ev

# start basic subsystems
export KAFKA_HEAP_OPTS="-Xmx100M -Xms100M"
/kafka/bin/zookeeper-server-start.sh -daemon /kafka/config/zookeeper.properties
/kafka/bin/kafka-server-start.sh -daemon /kafka/config/server.properties

rsyslogd -f /etc/rsyslogd.conf 
pg_ctlcluster 13 main start &
mongod --bind_ip 127.0.0.1  --logpath /logs/mongodb.log &
redis-server /etc/redis/redis.conf &
MYSQL_ROOT_PASSWORD="CGRateS.org" /scripts/mariadb-ep.sh mysqld > /logs/mariadb_script.log 2>&1
rabbitmq-server > /logs/rabbitmq.log  2>&1 &


START_TIMEOUT=600

start_timeout_exceeded=false
count=0
step=10
while netstat -lnt | awk '$4 ~ /:9092$/ {exit 1}'; do
    echo "waiting for kafka to be ready"
    sleep $step;
    count=$((count + step))
    if [ $count -gt $START_TIMEOUT ]; then
        start_timeout_exceeded=true
        break
    fi
done

if $start_timeout_exceeded; then
    echo "Not able to auto-create topic (waited for $START_TIMEOUT sec)"
    exit 1
fi

/kafka/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic cgrates
/kafka/bin/kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic cgrates_cdrs



gosu postgres psql  -c "CREATE USER cgrates password 'CGRateS.org';"  > /dev/null 2>&1
gosu postgres createdb -e -O cgrates cgrates > /dev/null 2>&1



PGPASSWORD="CGRateS.org" psql -U "cgrates" -h "localhost" -d cgrates -f /scripts/postgres/create_cdrs_tables.sql >/dev/null 2>&1
PGPASSWORD="CGRateS.org" psql -U "cgrates" -h "localhost" -d cgrates -f /scripts/postgres/create_tariffplan_tables.sql >/dev/null 2>&1


mongo --quiet /scripts/create_user.js >/dev/null 2>&1


mysql -u root -pCGRateS.org -h localhost < /scripts/mysql/create_db_with_users.sql > /dev/null 2>&1
mysql -u root -pCGRateS.org -h localhost < /scripts/mysql/create_db_with_users_extra.sql > /dev/null 2>&1
mysql -u root -pCGRateS.org -h localhost -D cgrates < /scripts/mysql/create_cdrs_tables.sql > /dev/null 2>&1
mysql -u root -pCGRateS.org -h localhost -D cgrates < /scripts/mysql/create_tariffplan_tables.sql > /dev/null 2>&1

ln -s /cgrates/data /usr/share/cgrates