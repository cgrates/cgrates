#! /usr/bin/env sh


if test $# -lt 2; then
        echo ""
        echo "setup_cgr_db.sh <db_user> <db_password> [<db_host>]"
        echo ""
        exit 0
fi

host=$3
if [ -z "$3" ]; then
	host="localhost" 
fi

mysql -u $1 -p$2 -h $host < create_db_with_users.sql
cu=$?

mysql -u $1 -p$2 -h $host -D cgrates < create_cdrs_tables.sql
cdrt=$?
mysql -u $1 -p$2 -h $host -D cgrates < create_costdetails_tables.sql
cdt=$?
mysql -u $1 -p$2 -h $host -D cgrates < create_mediator_tables.sql
mdt=$?
mysql -u $1 -p$2 -h $host -D cgrates < create_tariffplan_tables.sql
tpt=$?

if [ $cu = 0 ] && [ $cdrt = 0 ] && [ $cdt = 0 ] && [ $mdt = 0 ] && [ $tpt = 0 ]; then
	echo ""
	echo "\t+++ CGR-DB successfully set-up! +++"
	echo ""
	exit 0
fi


