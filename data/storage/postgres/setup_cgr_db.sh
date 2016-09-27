#! /usr/bin/env sh


user=$1
if [ -z "$1" ]; then
	user="cgrates"
fi

host=$2
if [ -z "$2" ]; then
	host="localhost"
fi

DIR="$(dirname "$(readlink -f "$0")")"

"$DIR"/create_db_with_users.sh

export PGPASSWORD="CGRateS.org"

psql -U $user -h $host -d cgrates -f "$DIR"/create_cdrs_tables.sql
cdrt=$?
psql -U $user -h $host -d cgrates -f "$DIR"/create_tariffplan_tables.sql
tpt=$?

if [ $cdrt = 0 ] && [ $tpt = 0 ]; then
	echo -e "\n\t+++ CGR-DB successfully set-up! +++\n"
	exit 0
fi


