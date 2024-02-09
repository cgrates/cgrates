.. _cgr-loader:

cgr-loader
----------

Tool used to load/import TariffPlan data into CGRateS databases.

Can be used to:
 * load TariffPlan data from **csv files** to **DataDB**.
 * import TariffPlan data from **csv files** to **StorDB** as offline data. ``-to_stordb -tpid``
 * import TariffPlan data from **StorDB** to **DataDB**. ``-from_stordb -tpid``

Customisable through the use of :ref:`JSON configuration <configuration>` or command line arguments (higher prio).


::

 $ cgr-loader -h
 Usage of cgr-loader:
  -api_key string
    	Api Key used to comosed ArgDispatcher
  -caches_address string
    	CacheS component to contact for cache reloads, empty to disable automatic cache reloads (default "*localhost")
  -caching string
    	Caching strategy used when loading TP
  -caching_delay duration
    	Adds delay before cache reload
  -config_path string
    	Configuration directory path.
  -datadb_host string
    	The DataDb host to connect to. (default "127.0.0.1")
  -datadb_name string
    	The name/number of the DataDb to connect to. (default "10")
  -datadb_passwd string
    	The DataDb user's password.
  -datadb_port string
    	The DataDb port to bind to. (default "6379")
  -datadb_type string
    	The type of the DataDB database <*redis|*mongo> (default "*redis")
  -datadb_user string
    	The DataDb user to sign in as. (default "cgrates")
  -dbdata_encoding string
    	The encoding used to store object data in strings (default "msgpack")
  -disable_reverse_mappings
    	Will disable reverse mappings rebuilding
  -dry_run
    	When true will not save loaded data to dataDb but just parse it for consistency and errors.
  -field_sep string
    	Separator for csv file (by default "," is used) (default ",")
  -flush_stordb
    	Remove tariff plan data for id from the database
  -from_stordb
    	Load the tariff plan from storDb to dataDb
  -import_id string
    	Uniquely identify an import/load, postpended to some automatic fields
  -mongoConnScheme string
    	Scheme for MongoDB connection <mongodb|mongodb+srv> (default "mongodb")
  -mongoQueryTimeout duration
    	The timeout for queries (default 10s)
  -path string
    	The path to folder containing the data files (default "./")
  -redisCACertificate string
    	Path to the CA certificate
  -redisClientCertificate string
    	Path to the client certificate
  -redisClientKey string
    	Path to the client key
  -redisCluster
    	Is the redis datadb a cluster
  -redisClusterOndownDelay duration
    	The delay before executing the commands if the redis cluster is in the CLUSTERDOWN state
  -redisClusterSync duration
    	The sync interval for the redis cluster (default 5s)
  -redisConnectAttempts int
    	The maximum amount of dial attempts (default 20)
  -redisConnectTimeout duration
    	The amount of wait time until timeout for a connection attempt
  -redisMaxConns int
    	The connection pool size (default 10)
  -redisReadTimeout duration
    	The amount of wait time until timeout for reading operations
  -redisSentinel string
    	The name of redis sentinel
  -redisTLS
    	Enable TLS when connecting to Redis
  -redisWriteTimeout duration
    	The amount of wait time until timeout for writing operations
  -remove
    	Will remove instead of adding data from DB
  -route_id string
    	RouteID used to comosed ArgDispatcher
  -rpc_encoding string
    	RPC encoding used <*gob|*json> (default "*json")
  -scheduler_address string
    	 (default "*localhost")
  -stordb_host string
    	The storDb host to connect to. (default "127.0.0.1")
  -stordb_name string
    	The name/number of the storDb to connect to. (default "cgrates")
  -stordb_passwd string
    	The storDb user's password. (default "CGRateS.org")
  -stordb_port string
    	The storDb port to bind to. (default "3306")
  -stordb_type string
    	The type of the storDb database <*mysql|*postgres|*mongo> (default "*mysql")
  -stordb_user string
    	The storDb user to sign in as. (default "cgrates")
  -tenant string
    	 (default "cgrates.org")
  -timezone string
    	Timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB> (default "Local")
  -to_stordb
    	Import the tariff plan from files to storDb
  -tpid string
    	The tariff plan ID from the database
  -verbose
    	Enable detailed verbose logging output
  -version
    	Prints the application version.
