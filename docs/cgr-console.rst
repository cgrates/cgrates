.. _cgr-console:

cgr-console
-----------

Command line tool used to interface via the APIs implemented within :ref:cgr-engine.

Configurable via command line arguments.

::

 $ cgr-console -help
 Usage of cgr-console:
  -ca_path string
    	path to CA for tls connection(only for self sign certificate)
  -connect_attempts int
    	Connect attempts (default 3)
  -connect_timeout int
    	Connect timeout in seconds  (default 1)
  -crt_path string
    	path to certificate for tls connection
  -key_path string
    	path to key for tls connection
  -max_reconnect_interval int
    	Maximum reconnect interval
  -reconnects int
    	Reconnect attempts (default 3)
  -reply_timeout int
    	Reply timeout in seconds  (default 300)
  -rpc_encoding string
    	RPC encoding used <*gob|*json> (default "*json")
  -server string
    	server address host:port (default "127.0.0.1:2012")
  -tls
    	TLS connection
  -verbose
    	Show extra info about command execution.
  -version
    	Prints the application version.



.. hint:: # cgr-console status