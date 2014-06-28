/etc/init.d/mysql start
/usr/share/cgrates/tutorials/fs_csv/freeswitch/etc/init.d/freeswitch start
mysqladmin -u root password CGRateS.org
cd /usr/share/cgrates/storage/mysql; ./setup_cgr_db.sh root CGRateS.org localhost
cd /
bash --rcfile /root/.bashrc
