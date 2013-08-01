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
git checkout-index -f --prefix=$PKG_DIR/ skel
chmod -R a-s $PKG_DIR/DEBIAN/
cp cp ../conf/cgrates.cfg $PKG_DIR/etc/cgrates/
cp -r ../../data/ $PKG_DIR/usr/share/cgrates/data
cp /usr/local/goapps/bin/cgr-* $PKG_DIR/usr/bin/
dpkg-deb --build $PKG_DIR

