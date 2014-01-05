#! /usr/bin/env sh

if [ -z "$1" ]
then
	echo ""
	echo "Package version not provided, exiting ..."
	echo ""
	exit 0
fi

WORK_DIR=/tmp/cgrates_$1
rm -rf $WORK_DIR
mkdir -p $WORK_DIR
cp -r debian $WORK_DIR/debian
cd $WORK_DIR
git clone https://github.com/cgrates/cgrates.git src/github.com/cgrates/cgrates
dpkg-buildpackage -us -uc
rm -rf $WORK_DIR
