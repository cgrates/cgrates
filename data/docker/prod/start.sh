/etc/init.d/mysql start
/etc/init.d/postgresql start
/etc/init.d/redis-server start

cd /usr/share/cgrates/storage/mysql && ./setup_cgr_db.sh root CGRateS.org
cd /usr/share/cgrates/storage/postgres && ./setup_cgr_db.sh

/usr/share/cgrates/tutorials/fs_evsock/freeswitch/etc/init.d/freeswitch start
 
# Docker doesn't have syslog. Let others modify this to send out logs if needed
sed -i 's/config_dir/config_path/g' /usr/share/cgrates/tutorials/fs_evsock/cgrates/etc/init.d/cgrates
sed -i 's/\/etc\/cgrates/\/etc\/cgrates -httprof_path=\/pprof -logger=*stdout/g' /usr/share/cgrates/tutorials/fs_evsock/cgrates/etc/init.d/cgrates

# Get our data ready
/usr/bin/cgr-migrator -exec=*set_versions -config_path=/usr/share/cgrates/tutorials/fs_evsock/cgrates/etc/cgrates/

# Let FreeSWITCH start up
sleep 5
/usr/share/cgrates/tutorials/fs_evsock/cgrates/etc/init.d/cgrates start

bash --rcfile /root/.bashrc
