#! /usr/bin/env sh
	echo ""
	echo "rc7_to_rc8.sh"
#settings

#DBs Config
datadb="redis"
stordb="mysql"
	echo "dataDB:"$datadb " storDB:"$stordb
	echo ""
#dataDBs
case $datadb in
"redis")
#Redis Config
export cgr_from_host='127.0.0.1' 
export cgr_from_port=6379
export cgr_from_db=11
export cgr_from_pass=''

export cgr_to_host='127.0.0.1'
export cgr_to_port=6379
export cgr_to_db=10
export cgr_to_pass='' # Not used
;;
"mongo")
#Mongo Config
export cgr_from_host='127.0.0.1'
export cgr_from_port='27017'
export cgr_from_db='11'
export cgr_from_auth_db='cgrates' # Auth db on source server
export cgr_from_user='cgrates'
export cgr_from_pass=''

export cgr_to_host='127.0.0.1'
export cgr_to_port='27017'
export cgr_to_db='10'
export cgr_to_auth_db="cgrates" # Auth db on target server
export cgr_to_user='cgrates'
export cgr_to_pass=''
;;
esac

export PGPASSWORD="CGRateS.org"
#StorDBs

case $stordb in

"mysql")
#mysql Config
user="cgrates"
host="127.0.0.1"
db="cgrates"
;;

"postgres")
#postgres Config
user="cgrates"
host="127.0.0.1"
db="cgrates"
;;
esac

DIR="$(dirname "$(readlink -f "$0")")"

#DataDB switch
case $datadb in 

"redis")
./dbsmerge_redis.py
;;

"mongo")
./dbsmerge_mongo.py
;;
esac

#StorDB switch
case $stordb in 

"mysql")
mysql -u$user -p$PGPASSWORD -h $host < "$DIR"/mysql_tables_update.sql
up=$?
mysql -u$user -p$PGPASSWORD -h $host -D cgrates < "$DIR"/mysql_cdr_migration.sql
mig=$?
;;

"postgres")
psql -U $user -h $host -d cgrates -f "$DIR"/pq_tables_update.sql
up=$?
psql -U $user -h $host -d cgrates -f "$DIR"/pg_cdr_migration.sql
mig=$?
;;
esac

if [ $up = 0 ] && [ $mig = 0 ]; then
	echo -e "\n\t+++ The script ran successfully ! +++\n"
	exit 0
fi