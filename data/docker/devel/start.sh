export GOROOT=/root/go
export GOPATH=/root/code
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

/etc/init.d/mysql start
/usr/share/cgrates/tutorials/fs_csv/freeswitch/etc/init.d/freeswitch start
mysqladmin -u root password CGRateS.org

# create a link to data dir
ln -s /root/code/src/github.com/cgrates/cgrates/data /usr/share/cgrates

# expand freeswitch json conf
tar xzf /usr/share/cgrates/tutorials/fs_json/freeswitch/etc/freeswitch_conf.tar.gz
    
# expand freeswitch csv 
tar xzf /usr/share/cgrates/tutorials/fs_csv/freeswitch/etc/freeswitch_conf.tar.gz
    
# create link to cgrates dir
ln -s /root/code/src/github.com/cgrates/cgrates /root/cgr

# create cgr-engine link
ln -s /root/code/bin/cgr-engine /usr/bin/cgr-engine

cd /usr/share/cgrates/storage/mysql; ./setup_cgr_db.sh root CGRateS.org localhost
cd /

cgr-engine -config /root/cgr/data/conf/cgrates.cfg &

bash --rcfile /root/.bashrc
