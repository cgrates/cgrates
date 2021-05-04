.. _cgr-tester:

cgr-tester
----------

Command line stress testing tool configurable via command line arguments.

::
 
 $ cgr-tester -h
 Usage of cgr-tester:
  -category string
    	The Record category to test. (default "call")
  -config_path string
    	Configuration directory path.
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
    	The type of the DataDb database <redis> (default "redis")
  -datadb_user string
    	The DataDb user to sign in as. (default "cgrates")
  -dbdata_encoding string
    	The encoding used to store object data in strings. (default "msgpack")
  -destination string
    	The destination to use in queries. (default "1002")
  -file_path string
    	read requests from file with path
  -json
    	Use JSON RPC
  -memprofile string
    	write memory profile to this file
  -parallel int
    	run n requests in parallel
  -rater_address string
    	Rater address for remote tests. Empty for internal rater.
  -redisSentinel string
    	The name of redis sentinel
  -redisCluster bool
    	Is the redis datadb a cluster
  -cluster_sync string
    	The sync interval for the redis cluster
  -cluster_ondown_delay string
    	The delay before executing the commands if thredis cluster is in the CLUSTERDOWN state
  -query_timeout string
    	The timeout for queries
  -req_separator string
    	separator for requests in file (default "\n\n")
  -runs int
    	stress cycle number (default 100000)
  -subject string
    	The rating subject to use in queries. (default "1001")
  -tenant string
    	The type of record to use in queries. (default "cgrates.org")
  -tor string
    	The type of record to use in queries. (default "*voice")
  -usage string
    	The duration to use in call simulation. (default "1m")
  -version
    	Prints the application version.
