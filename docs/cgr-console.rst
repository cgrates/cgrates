.. _cgr-console:
2.2. cgr-console
----------------
Command line tool used to interface with the APIs implemented within `cgr-engine`_.
Configurable via command line arguments.

::

 $ cgr-console -help
 Usage of cgr-console:
  -ca_path string
    	path to CA for tls connection(only for self sign certificate)
  -crt_path string
    	path to certificate for tls connection
  -key_path string
    	path to key for tls connection
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