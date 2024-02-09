.. _cgr-tester:

cgr-tester
----------

Command line stress testing tool configurable via command line arguments.

::
 
 $ cgr-tester -h
 Usage of cgr-tester:
  -calls int
    	run n number of calls (default 100)
  -category string
    	The Record category to test. (default "call")
  -config_path string
    	Configuration directory path.
  -cps int
    	run n requests in parallel (default 100)
  -cpuprofile string
    	write cpu profile to file
  -datadb_host string
    	The DataDb host to connect to. (default "127.0.0.1")
  -datadb_name string
    	The name/number of the DataDb to connect to. (default "10")
  -datadb_pass string
    	The DataDb user's password.
  -datadb_port string
    	The DataDb port to bind to. (default "6379")
  -datadb_type string
    	The type of the DataDb database <redis> (default "*redis")
  -datadb_user string
    	The DataDb user to sign in as. (default "cgrates")
  -dbdata_encoding string
    	The encoding used to store object data in strings. (default "msgpack")
  -destination string
    	The destination to use in queries. (default "1002")
  -digits int
    	Number of digits Account and Destination will have (default 10)
  -exec string
    	Pick what you want to test <*sessions|*cost>
  -file_path string
    	read requests from file with path
  -json
    	Use JSON RPC
  -max_usage duration
    	Maximum usage a session can have (default 5s)
  -memprofile string
    	write memory profile to this file
  -min_usage duration
    	Minimum usage a session can have (default 1s)
  -mongoConnScheme string
    	Scheme for MongoDB connection <mongodb|mongodb+srv> (default "mongodb")
  -mongoQueryTimeout duration
    	The timeout for queries (default 10s)
  -rater_address string
    	Rater address for remote tests. Empty for internal rater.
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
  -redisWriteTimeout duration
    	The amount of wait time until timeout for writing operations
  -req_separator string
    	separator for requests in file (default "\n\n")
  -request_type string
    	Request type of the call (default "*rated")
  -runs int
    	stress cycle number (default 100000)
  -subject string
    	The rating subject to use in queries. (default "1001")
  -tenant string
    	The type of record to use in queries. (default "cgrates.org")
  -timeout duration
    	After last call, time out after this much duration (default 10s)
  -tor string
    	The type of record to use in queries. (default "*voice")
  -update_interval duration
    	Time duration added for each session update (default 1s)
  -usage string
    	The duration to use in call simulation. (default "1m")
  -verbose
    	Enable detailed verbose logging output
  -version
    	Prints the application version.
