Tutorial
========

The general steps to get up and running with CGRateS are:

#. Create CSV files containing the initial data for CGRateS, see :ref:`data-importing`.
#. Load the data in the databases using the loader tool.
#. Start the balancer or rater and connect it to the call switch, see :ref:`running`.
#. Start one ore more raters.
#. Make API calls to the balancer/rater or just let the session manager do the work.

Instalation
-----------
**Using packages**

**Using source**

After the go environment is installed_ and setup_ just issue the following commands:
::

	go get github.com/cgrate/cgrates

This will install the sources and compile all available tools	
	
.. _installed: http://golang.org/doc/install
.. _setup: http://golang.org/doc/code.html


Running
-------

There are only three main command to used with CGRateS:

cgr-balancer
	The cgr-balancer will open a JSON RPC server and an HTTP server ready for taking external requests. It will also open a rater server on witch the raters will register themselves when they start.
::

	rif@grace:~$ cgr-balancer --help
	Usage of cgr-balancer:
	  -freeswitchpass="ClueCon": freeswitch address host:port
	  -freeswitchsrv="localhost:8021": freeswitch address host:port
	  -httpapiaddr="127.0.0.1:8000": Http API server address (localhost:2002)
	  -json=false: use JSON for RPC encoding
	  -jsonrpcaddr="127.0.0.1:2001": Json RPC server address (localhost:2001)
	  -rateraddr="127.0.0.1:2000": Rater server address (localhost:2000)


cgr-rater
	The cgr-rater can be provided with the balancer server address and can be configured to listen to a specific interface and port.
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

cgr-console
	The cgr-console is a command line tool used to access the balancer (or the rater directly) to call all the API methods offered by CGRateS.
::

	rif@grace:~$ cgr-console --help
	Usage of cgr-console:
	  -amount=100: Amount for different operations
	  -balancer="127.0.0.1:2001": balancer address host:port
	  -cstmid="vdf": Customer identification
	  -dest="0256": Destination prefix
	  -subject="rif": The client who made the call
	  -te="2012-02-09T00:10:00Z": Time end
	  -tor=0: Type of record
	  -ts="2012-02-09T00:00:00Z": Time start

cgr-loader
	The loader is the most configurable tool because it has options for each of the three supported databases (kyoto, redis and mongodb).
	Apart from that multi-database options it is quite easy to be used.
	The apfile, destfile, tpfile and ubfile parameters are for specifying the input json files.
	The storage parameter specifies the database to be used and then the databases access information (host:port or file) has to be provided.

	:Example: cgr-loader -storage=kyoto -kyotofile=storage.kch -apfile=activationperiods.json -destfile=destinations.json -tpfile=tariffplans.json -ubfile=userbudgets.json
::

	rif@grace:~$ cgr-loader --help
	Usage of cgr-loader:
	  -accountactions="AccountActions.csv": Account actions file
	  -actions="Actions.csv": Actions file
	  -actiontimings="ActionTimings.csv": Actions timings file
	  -actiontriggers="ActionTriggers.csv": Actions triggers file
	  -destinations="Destinations.csv": Destinations file
	  -flush=false: Flush the database before importing
	  -month="Months.csv": Months file
	  -monthdays="MonthDays.csv": Month days file
	  -pass="": redis database password
	  -rates="Rates.csv": Rates file
	  -ratetimings="RateTimings.csv": Rates timings file
	  -ratingprofiles="RatingProfiles.csv": Rating profiles file
	  -rdb=10: redis database number (10)
	  -redisserver="127.0.0.1:6379": redis server address (tcp:127.0.0.1:6379)
	  -separator=",": Default field separator
	  -timings="Timings.csv": Timings file
	  -weekdays="WeekDays.csv": Week days file


rif@grace:~$ cgr-balancer --help
Usage of cgr-balancer:
  -freeswitchpass="ClueCon": freeswitch address host:port
  -freeswitchsrv="localhost:8021": freeswitch address host:port
  -httpapiaddr="127.0.0.1:8000": Http API server address (localhost:2002)
  -json=false: use JSON for RPC encoding
  -jsonrpcaddr="127.0.0.1:2001": Json RPC server address (localhost:2001)
  -rateraddr="127.0.0.1:2000": Rater server address (localhost:2000)

rif@grace:~$ cgr-sessionmanager --help
Usage of cgr-sessionmanager:
  -balancer="127.0.0.1:2000": balancer address host:port
  -freeswitchpass="ClueCon": freeswitch address host:port
  -freeswitchsrv="localhost:8021": freeswitch address host:port
  -json=false: use JSON for RPC encoding
  -redisdb=10: redis database number
  -redissrv="127.0.0.1:6379": redis address host:port
  -standalone=false: run standalone (run as a rater)

rif@grace:~$ cgr-mediator --help
Usage of cgr-mediator:
  -dbname="cgrates": The name of the database to connect to.
  -freeswitchcdr="Master.csv": Freeswitch Master CSV CDR file.
  -host="localhost": The host to connect to. Values that start with / are for unix domain sockets.
  -password="": The user's password.
  -port="5432": The port to bind to.
  -resultfile="out.csv": Generated file containing CDR and price info.
  -user="": The user to sign in as.