#! /usr/bin/env sh


if test $# -lt 2; then
        echo ""
        echo "setup_cgr_db.sh <db_user> <db_password> [<db_host>]"
        echo ""
        exit 0
fi

user=$1
pass=$2
host=$3
if [ -z "$3" ]; then
	host="localhost"
fi
cqlsh -u $user -p $pass $host -f create_db_with_users.cql
cu=$?
cqlsh -u $user -p $pass $host -f create_cdrs_tables.cql
cdrt=$?

if [ $cu = 0 ] && [ $cdrt = 0 ]; then
	echo ""
	echo "\t+++ CGR-DB successfully set-up! +++"
	echo ""
	exit 0
fi
