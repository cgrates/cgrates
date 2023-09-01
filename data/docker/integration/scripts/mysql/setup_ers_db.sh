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

mysql -u $1 -p$2 -h $host < "$DIR"/create_ers_db.sql
cu=$?

if [ $cu = 0 ]; then
	echo "\n\t+++ CGR-DB successfully set-up! +++\n"
	exit 0
fi


