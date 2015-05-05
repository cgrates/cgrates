export GOROOT=/root/go
export GOPATH=/root/code
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

/etc/init.d/mysql start
/etc/init.d/postgresql start
/etc/init.d/redis-server start

# create a link to data dir
ln -s /root/code/src/github.com/cgrates/cgrates/data /usr/share/cgrates
# create link to cgrates dir
ln -s /root/code/src/github.com/cgrates/cgrates /root/cgr

cd /usr/share/cgrates/storage/mysql && ./setup_cgr_db.sh root CGRateS.org
cd /usr/share/cgrates/storage/postgres && ./setup_cgr_db.sh

# build and install cgrates
/root/cgr/update_external_libs.sh
go install github.com/cgrates/cgrates

# create cgr-engine link
ln -s /root/code/bin/cgr-engine /usr/bin/cgr-engine

# expand freeswitch conf
cd /usr/share/cgrates/tutorials/fs_evsock/freeswitch/etc/ && tar xzf freeswitch_conf.tar.gz       


bash --rcfile /root/.bashrc
