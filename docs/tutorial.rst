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

The CGRateS suite is formed by seven tools described bellow.

cgr-balancer
~~~~~~~~~~~~
The cgr-balancer will open a JSON RPC server and an HTTP server ready for taking external requests. It will also open a rater server on witch the raters will register themselves when they start.
::

	rif@grace:~$ cgr-balancer --help
	Usage of cgr-balancer:
	  -freeswitch=false: connect to freeswitch server
	  -freeswitchpass="ClueCon": freeswitch address host:port
	  -freeswitchsrv="localhost:8021": freeswitch address host:port
	  -httpapiaddr="127.0.0.1:8000": Http API server address (localhost:2002)
	  -json=false: use JSON for RPC encoding
	  -jsonrpcaddr="127.0.0.1:2001": Json RPC server address (localhost:2001)
	  -rateraddr="127.0.0.1:2000": Rater server address (localhost:2000)

:Example: cgr-balancer -freeswitch=true -httpapiaddr=127.0.0.1:6060 -json=true

cgr-rater
~~~~~~~~~
The cgr-rater can be provided with the balancer server address and can be configured to listen to a specific interface and port. It is an auxiliary tool only and is meant to be used for housekeeping duties (better alternative to curl inspection).
::

	rif@grace:~$ cgr-rater --help
	Usage of cgr-rater:
	  -balancer="127.0.0.1:2000": balancer address host:port
	  -freeswitch=false: connect to freeswitch server
	  -freeswitchpass="ClueCon": freeswitch address host:port
	  -freeswitchsrv="localhost:8021": freeswitch address host:port
	  -json=false: use JSON for RPC encoding
	  -listen="127.0.0.1:1234": listening address host:port
	  -redisdb=10: redis database number
	  -redissrv="127.0.0.1:6379": redis address host:port
	  -standalone=false: start standalone server (no balancer)

:Example: cgr-rater -balancer=127.0.0.1:2000

cgr-console
~~~~~~~~~~~
The cgr-console is a command line tool used to access the balancer (or the rater directly) to call all the API methods offered by CGRateS. It is
::

	rif@grace:~$ cgr-console 
	List of commands:
	        getcost
	        getmaxsessiontime
	        debitbalance
	        debitsms
	        debitseconds
	        resetuserbudget
	        status
	  -direction="OUT": Call direction
	  -tenant="vdf": Tenant identificator
	  -tor="0": Type of record
	  -amount=100: Amount for different operations
	  -dest="041": Call destination	  
	  -server="127.0.0.1:2001": server address host:port
	  -subject="rif": The client who made the call
	  -account="rif": The the user balance to be used
	  -start="2012-02-09T00:00:00Z": Time start
	  -end="2012-02-09T00:10:00Z": Time end	  

:Example: cgr-console getcost -subject=rif -dest=0723045326 -start=2012-07-13T15:38:00Z -end=2012-07-13T15:39:00Z

cgr-loader
~~~~~~~~~~

This tool is used for importing the data from CSV files into the CGRateS database system. The structure of the CSV files is described in the :ref:`data-importing` chapter.

::

	rif@grace:~$ cgr-loader --help
	Usage of cgr-loader:
	  -accountactions="": Account actions file
	  -actions="": Actions file
	  -actiontimings="": Actions timings file
	  -actiontriggers="": Actions triggers file
	  -destinations="": Destinations file
	  -flush=false: Flush the database before importing
	  -month="": Months file
	  -monthdays="": Month days file
	  -pass="": redis database password
	  -rates="": Rates file
	  -ratetimings="": Rates timings file
	  -ratingprofiles="": Rating profiles file
	  -redisdb=10: redis database number (10)
	  -redissrv="127.0.0.1:6379": redis server address (tcp:127.0.0.1:6379)
	  -separator=",": Default field separator
	  -timings="": Timings file
	  -weekdays="": Week days file

:Example: cgr-loader -destinations=Destinations.csv

cgr-sessionmanager
~~~~~~~~~~~~~~~~~~

Session manager connects and monitors the freeswitch server issuing API request to other CGRateS components. It can run in standalone mode for minimal system configuration. It logs the calls information to a postgres database in order to be used by the mediator tool.

::

	rif@grace:~$ cgr-sessionmanager --help
	Usage of cgr-sessionmanager:
	  -balancer="127.0.0.1:2000": balancer address host:port
	  -freeswitchpass="ClueCon": freeswitch address host:port
	  -freeswitchsrv="localhost:8021": freeswitch address host:port
	  -json=false: use JSON for RPC encoding
	  -redisdb=10: redis database number
	  -redissrv="127.0.0.1:6379": redis address host:port
	  -standalone=false: run standalone (run as a rater)

:Example: cgr-sessionmanager -standalone=true

cgr-mediator
~~~~~~~~~~~~

The mediator parses the call logs written in a postgres database by the session manager and writes the call costs to a freeswitch CDR file.

The structure of the table (as an SQL command) is the following::

	CREATE TABLE callcosts (
	uuid varchar(80) primary key,direction varchar(32),
	tenant varchar(32),tor varchar(32),
	subject varchar(32),
	account varchar(32),
	destination varchar(32),
	cost real,
	conect_fee real,
	timespans text
	);

::

	rif@grace:~$ cgr-mediator --help
	Usage of cgr-mediator:
	  -dbname="cgrates": The name of the database to connect to.
	  -freeswitchcdr="Master.csv": Freeswitch Master CSV CDR file.
	  -host="localhost": The host to connect to. Values that start with / are for UNIX domain sockets.
	  -password="": The user's password.
	  -port="5432": The port to bind to.
	  -resultfile="out.csv": Generated file containing CDR and price info.
	  -user="": The user to sign in as.

:Example: cgr-mediator -freeswitchcdr="logs.csv"

cgr-scheduler
~~~~~~~~~~~~~

The scheduler is loading the timed actions form database and executes them as appropriate, It will execute all run once actions as they are loaded. It will reload all the action timings from the database when it received system HUP signal (pkill -1 cgr-schedule).

::

	rif@grace:~$ cgr-scheduler --help
	Usage of cgr-scheduler:
	  -pass="": redis database password
	  -rdb=10: redis database number (10)
	  -redisserver="127.0.0.1:6379": redis server address (tcp:127.0.0.1:6379)	  

:Example: cgr-scheduler -rdb=2 -pass="secret"