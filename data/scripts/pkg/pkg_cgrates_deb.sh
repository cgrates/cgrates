#!/bin/bash

WORK_DIR=/tmp/cgrates
rm -rf $WORK_DIR
mkdir -p $WORK_DIR
cp -r debian $WORK_DIR/debian
cd $WORK_DIR
git clone https://github.com/cgrates/cgrates.git src/github.com/cgrates/cgrates
tar xvzf src/github.com/cgrates/cgrates/data/tutorials/fs_evsock/freeswitch/etc/freeswitch_conf.tar.gz -C src/github.com/cgrates/cgrates/data/tutorials/fs_evsock/freeswitch/etc/
rm src/github.com/cgrates/cgrates/data/tutorials/fs_evsock/freeswitch/etc/freeswitch_conf.tar.gz
dpkg-buildpackage -us -uc
#rm -rf $WORK_DIR
