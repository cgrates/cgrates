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
cp ../conf/cgrates.cfg $PKG_DIR/etc/cgrates/cgrates.cfg
cp -r ../../data/ $PKG_DIR/usr/share/cgrates/
cp /usr/local/goapps/bin/cgr-* $PKG_DIR/usr/bin/
dpkg-deb --build $PKG_DIR

