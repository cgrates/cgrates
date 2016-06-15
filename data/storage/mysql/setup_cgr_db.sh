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

DIR="$(dirname "$(readlink -f "$0")")"

mysql -u $1 -p$2 -h $host < "$DIR"/create_db_with_users.sql
cu=$?
mysql -u $1 -p$2 -h $host -D cgrates < "$DIR"/create_cdrs_tables.sql
cdrt=$?
mysql -u $1 -p$2 -h $host -D cgrates < "$DIR"/create_tariffplan_tables.sql
tpt=$?

if [ $cu = 0 ] && [ $cdrt = 0 ] && [ $tpt = 0 ]; then
	echo -e "\n\t+++ CGR-DB successfully set-up! +++\n"
	exit 0
fi


