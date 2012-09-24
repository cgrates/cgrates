Tutorial
========

The general steps to get up and running with CGRateS are:

#. Create CSV files containing the initial data for CGRateS, see :ref:`data-importing`.
#. Load the data in the databases using the loader tool.
#. Start the balancer or rater and connect it to the call switch, see Running_.
#. Start one ore more raters.
#. Make API calls to the balancer/rater or just let the session manager do the work.

Instalation
-----------
Using packages
~~~~~~~~~~~~~~

Using source
~~~~~~~~~~~~

After the go environment is installed_ and setup_ just issue the following commands:
::

	go get github.com/cgrates/cgrates

This will install the sources and compile all available tools	
	
.. _installed: http://golang.org/doc/install
.. _setup: http://golang.org/doc/code.html


Running
-------

The CGRateS suite is formed by three tools described bellow. We'll start with the most important one, cgr-rater which is configured with an ini style configuration file.


cgr-rater
~~~~~~~~~
The cgr-rater can be provided with the balancer server address and can be configured to listen to a specific interface and port. It is an auxiliary tool only and is meant to be used for housekeeping duties (better alternative to curl inspection).
::

   rif@grace:~$ cgr-rater -help
   Usage of cgr-rater:
      -config="rater_standalone.config": Configuration file location.
      -version=false: Prints the application version.


:Example: cgr-rater -config=full.config

Bellow there is a full configuration file:

::

   [global]
   datadb_type = redis # The main database: redis|mongo|postgres.
   datadb_host = 127.0.0.1:6379 # The host to connect to. Values that start with / are for UNIX domain sockets.
   datadb_name = 10 # The name of the database to connect to.
   logdb_type = mongo # The logging database: redis|mongo|postgres|same.
   logdb_host = localhost # The host to connect to. Values that start with / are for UNIX domain sockets.
   logdb_name = cgrates # The name of the database to connect to.

   [balancer]
   enabled = false # Start balancer server
   listen = 127.0.0.1:2001 # Balancer listen interface
   rpc_encoding = gob # Use json or gob for RPC encoding

   [rater]
   enabled = true # Start the rating service
   listen = 127.0.0.1:2001 # Listening address host:port, internal for internal communication only
   balancer = disabled # If defined it will register to balancer as worker
   rpc_encoding = gob # Use json or gob for RPC encoding

   [mediator]
   enabled = true # Start the mediator service
   cdr_path = /var/log/freeswitch # Freeswitch Master CSV CDR path
   cdr_out_path = /var/log/freeswitch/out # Freeswitch Master CSV CDR path
   rater = internal # Address where to access rater. Can be internal, direct rater address or the address of a balancer
   rpc_encoding = gob # Use json or gob for RPC encoding
   skipdb = true # Do not look in the database for logged cdrs, ask rater directly

   [scheduler]
   enabled = true # Start the schedule service

   [session_manager]
   enabled = true # Start the session manager service
   switch_type = freeswitch # The switch type to be used
   debit_period = 10 # The number of seconds to be debited in advance during a call
   rater = 127.0.0.1:2000 # Address where to access rater. Can be internal, direct rater address or the address of a balancer
   rpc_encoding = gob # Use json or gob for RPC encoding

   [freeswitch]
   server = localhost:8021 # Freeswitch address host:port
   pass = ClueCon # Freeswtch address host:port
   direction_index = 0
   tor_index = 1
   tenant_index = 2
   subject_index = 3
   account_index = 4
   destination_index = 5
   time_start_index = 6
   time_end_index = 7
   

There are various sections in the configuration file that define various services that the cgr-rater process can provide. If you are not interested in a certain service you can either leave it in the configuration with the enabled option set to false or remove the section entirely to reduce clutter.

The global sections define the databases to be used with used by CGRateS. The second database is used for logging the debit operations and various acctions operated on the accounts. The two databases can be the same type or different types. Currently we sopport redis, mongo and postgres (work in progress).

The balancer will open a JSON RPC server and an HTTP server ready for taking external requests. It will also open a rater server on witch the raters will register themselves when they start.

Session manager connects and monitors the freeswitch server issuing API request to other CGRateS components. It can run in standalone mode for minimal system configuration. It logs the calls information to a postgres database in order to be used by the mediator tool.

The scheduler is loading the timed actions form database and executes them as appropriate, It will execute all run once actions as they are loaded. It will reload all the action timings from the database when it received system HUP signal (pkill -1 cgr-rater).

The mediator parses the call logs written in the logging database by the session manager and writes the call costs to a freeswitch CDR file.

The structure of the table (as an SQL command) is the following::
::

	CREATE TABLE callcosts (
	uuid varchar(80) primary key,
   source varchar(32),
   direction varchar(32),
	tenant varchar(32),
   tor varchar(32),
	subject varchar(32),
	account varchar(32),
	destination varchar(32),
	cost real,
	conect_fee real,
	timespans text
	);



cgr-loader
~~~~~~~~~~

This tool is used for importing the data from CSV files into the CGRateS database system. The structure of the CSV files is described in the :ref:`data-importing` chapter.

::

   rif@grace:~$ cgr-loader -help
   Usage of cgr-loader:
      -dbhost="localhost": The database host to connect to.
      -dbname="10": he name/number of the database to connect to.
      -dbpass="": The database user's password.
      -dbport="6379": The database port to bind to.
      -dbtype="redis": The type of the database (redis|mongo|postgres)
      -dbuser="": The database user to sign in as.
      -flush=false: Flush the database before importing
      -path=".": The path containing the data files
      -version=false: Prints the application version.
   

:Example: cgr-loader -flush


cgr-console
~~~~~~~~~~~
The cgr-console is a command line tool used to access the balancer (or the rater directly) to call all the API methods offered by CGRateS. It is
::

   cgrrif@grace:~$ cgr-console -help
   Usage of cgr-console:
      -account="": The the user balance to be used
      -amount=0: Amount for different operations
      -cmd="": server address host:port
      -dest="": Call destination
      -direction="OUT": Call direction
      -end="": Time end (format: 2012-02-09T00:00:00Z)
      -json=false: Use JSON for RPC encoding.
      -server="127.0.0.1:2001": server address host:port
      -start="": Time start (format: 2012-02-09T00:00:00Z)
      -subject="": The client who made the call
      -tenant="": Tenant identificator
      -tor="0": Type of record
      -version=false: Prints the application version.

:Example: cgr-console -cmd=getcost -subject=rif -tenant=vdf -dest=419 -start=2012-02-09T00:00:00Z -end=2012-02-09T00:01:00Z

List of commands:
 - getcost
 - debit
 - maxdebit
 - getmaxsessiontime
 - debitbalance
 - debitsms
 - debitseconds
 - addrecievedcallseconds
 - flushcache
 - status
 - shutdown
