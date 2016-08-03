# edit servers config files
sed -i 's/127.0.0.1/0.0.0.0/g' /etc/redis/redis.conf /etc/mysql/my.cnf

/etc/init.d/rsyslog start
/etc/init.d/mysql start
/etc/init.d/redis-server start

DISABLE_AUTO_UPDATE="true" zsh
