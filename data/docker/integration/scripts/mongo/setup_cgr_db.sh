#! /usr/bin/env sh


mongo --quiet create_user.js
cu=$?

if [ $cu = 0 ]; then
	echo ""
	echo "\t+++ CGR-DB successfully set-up! +++"
	echo ""
	exit 0
fi


