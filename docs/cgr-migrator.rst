.. _cgr-migrator:

cgr-migrator
------------

Command line migration tool.

Customisable through the use of :ref:`JSON configuration <configuration>` or command line arguments (higher prio).

::

 $ cgr-migrator -h
 Usage of cgr-migrator:
  -config_path string
    	Configuration directory path.
  -datadb_host string
    	the DataDB host (default "127.0.0.1")
  -datadb_name string
    	the name/number of the DataDB (default "10")
  -datadb_passwd string
    	the DataDB password
  -datadb_port string
    	the DataDB port (default "6379")
  -datadb_type string
    	the type of the DataDB Database <*redis|*mongo> (default "redis")
  -datadb_user string
    	the DataDB user (default "cgrates")
  -dbdata_encoding string
    	the encoding used to store object Data in strings (default "msgpack")
  -dry_run
    	parse loaded data for consistency and errors, without storing it
  -exec string
    	fire up automatic migration <*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups|*filters|*stordb|*datadb>
  -out_datadb_host string
    	output DataDB host to connect to (default "*datadb")
  -out_datadb_name string
    	output DataDB name/number (default "*datadb")
  -out_datadb_passwd string
    	output DataDB password (default "*datadb")
  -out_datadb_port string
    	output DataDB port (default "*datadb")
  -out_datadb_type string
    	output DataDB type <*redis|*mongo> (default "*datadb")
  -out_datadb_user string
    	output DataDB user (default "*datadb")
  -out_dbdata_encoding string
    	the encoding used to store object Data in strings in move mode (default "*datadb")
  -out_redis_sentinel string
    	the name of redis sentinel (default "*datadb")
  -out_stordb_host string
    	output StorDB host (default "*stordb")
  -out_stordb_name string
    	output StorDB name/number (default "*stordb")
  -out_stordb_passwd string
    	output StorDB password (default "*stordb")
  -out_stordb_port string
    	output StorDB port (default "*stordb")
  -out_stordb_type string
    	output StorDB type for move mode <*mysql|*postgres|*mongo> (default "*stordb")
  -out_stordb_user string
    	output StorDB user (default "*stordb")
  -redisSentinel string
    	the name of redis sentinel
  -redisCluster bool
    	Is the redis datadb a cluster
  -cluster_sync string
    	The sync interval for the redis cluster
  -cluster_ondown_delay string
    	The delay before executing the commands if thredis cluster is in the CLUSTERDOWN state
  -query_timeout string
    	The timeout for queries
  -stordb_host string
    	the StorDB host (default "127.0.0.1")
  -stordb_name string
    	the name/number of the StorDB (default "cgrates")
  -stordb_passwd string
    	the StorDB password
  -stordb_port string
    	the StorDB port (default "3306")
  -stordb_type string
    	the type of the StorDB Database <*mysql|*postgres|*mongo> (default "mysql")
  -stordb_user string
    	the StorDB user (default "cgrates")
  -verbose
    	enable detailed verbose logging output
  -version
    	prints the application version
