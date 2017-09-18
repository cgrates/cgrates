#! /usr/bin/env sh
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
#Mongo Config //NOT SUPPORTED IN RC7
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
port="3306"
;;

"postgres")
#postgres Config
user="cgrates"
host="127.0.0.1"
db="cgrates"
port="5432"
;;
esac

DIR="$(dirname "$(readlink -f "$0")")"

#DataDB switch
case $datadb in 

"redis")
echo 'Calling script: dbsmerge_redis.py'
./dbsmerge_redis.py 
echo 'done!'
echo
;;

"mongo")
echo 'Calling script: dbsmerge_mongo.py'
./dbsmerge_mongo.py
echo 'done!'
echo
;;
esac

#StorDB switch
case $stordb in 

"mysql")
echo "Calling script: mysql_tables_update.sql"
mysql -u$user -p$PGPASSWORD -h $host < "$DIR"/mysql_tables_update.sql
up=$?
echo "done!"
echo
echo "Calling script: mysql_cdr_migration.sql"
mysql -u$user -p$PGPASSWORD -h $host -D cgrates < "$DIR"/mysql_cdr_migration.sql
mig=$?
echo "done!"
echo
echo 'Executing command cgr-migrator -migrate="*set_versions"'
cgr-migrator -datadb_host=$cgr_from_host -datadb_name=$cgr_to_db -datadb_passwd=$cgr_from_pass -datadb_port=$cgr_from_port -datadb_type=$datadb -stordb_host=$host -stordb_name=$user -stordb_passwd=$PGPASSWORD -stordb_port=$port  -stordb_type=$stordb -stordb_user=$user -migrate="*set_versions"
echo
echo 'Setting version for CostDetails'
mysql -u$user -p$PGPASSWORD -h $host -D cgrates < "$DIR"/set_version.sql
;;

"postgres")
echo "Calling script: pg_tables_update.sql"
psql -U $user -h $host -d cgrates -f "$DIR"/pg_tables_update.sql
up=$?
echo "done!"
echo
echo "Calling script: pg_cdr_migration.sql"
 psql -U $user -h $host -d cgrates -f "$DIR"/pg_cdr_migration.sql
mig=$?
echo "done!"
echo
echo 'Executing command cgr-migrator -migrate="*set_versions"'
cgr-migrator -datadb_host=$cgr_from_host -datadb_name=$cgr_to_db -datadb_passwd=$cgr_from_pass -datadb_port=$cgr_from_port -datadb_type=$datadb -stordb_host=$host -stordb_name=$user -stordb_passwd=$PGPASSWORD -stordb_port=$port  -stordb_type=$stordb -stordb_user=$user -migrate="*set_versions"
echo
echo 'Setting version for CostDetails'
 psql -U $user -h $host -d cgrates -f "$DIR"/set_version.sql
;;
esac

echo 'Executing command cgr-migrator -migrate="*cost_details,*accounts,*actions,*action_triggers,*action_plans,*shared_groups,*set_versions"'
cgr-migrator -datadb_host=$cgr_from_host -datadb_name=$cgr_to_db -datadb_passwd=$cgr_from_pass -datadb_port=$cgr_from_port -datadb_type=$datadb -stordb_host=$host -stordb_name=$user -stordb_passwd=$PGPASSWORD -stordb_port=$port  -stordb_type=$stordb -stordb_user=$user -verbose=true -stats=true -migrate="*cost_details,*accounts,*actions,*action_triggers,*action_plans,*shared_groups,*set_versions"