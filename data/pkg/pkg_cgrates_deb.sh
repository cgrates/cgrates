#! /usr/bin/env sh

if [ -z "$1" ]
then
	echo ""
	echo "Package version not provided, exiting ..."
	echo ""
	exit 0
fi

PKG_DIR=/tmp/cgrates_$1
rm -rf $PKG_DIR
cp -r skel $PKG_DIR
chmod -R a-s $PKG_DIR/DEBIAN/
mkdir -p $PKG_DIR/etc/cgrates
cp ../conf/cgrates.cfg $PKG_DIR/etc/cgrates/cgrates.cfg
mkdir -p $PKG_DIR/usr/share/cgrates
cp -r ../../data/ $PKG_DIR/usr/share/cgrates/
mkdir -p $PKG_DIR/usr/bin
cp /usr/local/goapps/bin/cgr-* $PKG_DIR/usr/bin/
mkdir -p $PKG_DIR/var/log/cgrates/cdr/cdrc/in
mkdir -p $PKG_DIR/var/log/cgrates/cdr/cdrc/out
mkdir -p $PKG_DIR/var/log/cgrates/cdr/cdrexport/csv
mkdir -p $PKG_DIR/var/log/cgrates/history
dpkg-deb --build $PKG_DIR

