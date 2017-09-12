#! /usr/bin/env sh

#settings

	echo ""
	echo "rc7_to_rc8.sh"
	echo ""

$datadb="redis"

if [$datadb="redis"];then
#Redis Config
export from_host   = '192.168.100.40'
export from_port   = 6379
export from_db     = 11
export from_pass   = ''

export to_host     = '192.168.100.40'
export to_port     = 6379
export to_db       = 10
export to_pass     = '' # Not used

else if [$datadb="mongo"];then
#Mongo Config
export from_host    = '127.0.0.1'
export from_port    = '27017'
export from_db      = '11'
export from_auth_db = 'cgrates' # Auth db on source server
export from_user    = 'cgrates'
export from_pass    = ''

export to_host      = '127.0.0.1'
export to_port      = '27017'
export to_db        = '10'
export to_auth_db   = "cgrates" # Auth db on target server
export to_user      = 'cgrates'
export to_pass      = ''
fi




if [$stordb="redis"];then
#Redis Config
export from_host   = '192.168.100.40'
export from_port   = 6379
export from_db     = 11
export from_pass   = ''

export to_host     = '192.168.100.40'
export to_port     = 6379
export to_db       = 10
export to_pass     = '' # Not used

else if [$stordb="mysql"];then
#Mongo Config
export from_host    = '127.0.0.1'
export from_port    = '27017'
export from_db      = '11'
export from_auth_db = 'cgrates' # Auth db on source server
export from_user    = 'cgrates'
export from_pass    = ''

export to_host      = '127.0.0.1'
export to_port      = '27017'
export to_db        = '10'
export to_auth_db   = "cgrates" # Auth db on target server
export to_user      = 'cgrates'
export to_pass      = ''
fi



DIR="$(dirname "$(readlink -f "$0")")"


case $datadb in 
"redis")
./dbsmerge_redis.py
;;
"mongo")
./dbsmerge_mongo.py
;;
esac

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
	echo -e "\n\t+++ CGR-DB successfully set-up! +++\n"
	exit 0
fi


