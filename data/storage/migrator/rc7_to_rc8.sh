#! /usr/bin/env sh

if test $# -lt 2; then
        echo ""
        echo "setup_cgr_db.sh <datadb_name> <stordb_name> <User> <host>"
        echo ""
        exit 0
fi


user=$3
if [ -z "$3" ]; then
	user="cgrates"
fi

host=$4
if [ -z "$4" ]; then
	host="localhost"
fi
export PGPASSWORD="CGRateS.org"

DIR="$(dirname "$(readlink -f "$0")")"


case $1 in 
"redis")
./dbsmerge_redis.py
;;
"mongo")
./dbsmerge_mongo.py
;;
esac

case $2 in 
	"mysql")
mysql -u$user -p$PGPASSWORD -h $host < "$DIR"/mysql_tables_update.sql
up=$?
mysql -u$user -p$PGPASSWORD -h $host -D cgrates < "$DIR"/mysql_cdr_migration.sql
mig=$?
#./usage_mysql.py What's the point of those changes?
;;
"postgres")
psql -U $user -h $host -d cgrates -f "$DIR"/pq_tables_update.sql
up=$?
psql -U $user -h $host -d cgrates -f "$DIR"/pg_cdr_migration.sql
mig=$?
#./usage_postgres.py What's the point of those changes?
;;
esac

if [ $up = 0 ] && [ $mig = 0 ]; then
	echo -e "\n\t+++ CGR-DB successfully set-up! +++\n"
	exit 0
fi


