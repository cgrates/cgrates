# edit servers config files
sed -i 's/127.0.0.1/0.0.0.0/g' /etc/redis/redis.conf /etc/mysql/my.cnf

# start services
service rsyslog start
service mysql start
service redis-server start
/root/code/bin/cgr-engine -config_path /root/cgr/data/conf/samples/osips_training

# setup mysql
cd /usr/share/cgrates/storage/mysql && ./setup_cgr_db.sh root CGRateS.org

# load tariff plan data
#cd /root/cgr/data/tariffplans/osips_training; cgr-loader

cd /root/cgr
DISABLE_AUTO_UPDATE="true" zsh
