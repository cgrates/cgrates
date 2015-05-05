/etc/init.d/mysql start
/etc/init.d/postgresql start
/etc/init.d/redis-server start

cd /usr/share/cgrates/storage/mysql && ./setup_cgr_db.sh root CGRateS.org
cd /usr/share/cgrates/storage/postgres && ./setup_cgr_db.sh

/usr/share/cgrates/tutorials/fs_evsock/freeswitch/etc/init.d/freeswitch start
/usr/share/cgrates/tutorials/fs_evsock/cgrates/etc/init.d/cgrates start

bash --rcfile /root/.bashrc
