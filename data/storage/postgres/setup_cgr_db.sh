#! /usr/bin/env sh


user=$1
if [ -z "$1" ]; then
	user="postgres" 
fi

host=$2
if [ -z "$2" ]; then
	host="localhost" 
fi

./create_db_with_users.sh

sudo -u $user psql -d cgrates -f create_cdrs_tables.sql
cdrt=$?
sudo -u $user psql -d cgrates -f create_tariffplan_tables.sql
tpt=$?

if [ $cdrt = 0 ] && [ $tpt = 0 ]; then
	echo ""
	echo "\t+++ CGR-DB successfully set-up! +++"
	echo ""
	exit 0
fi


