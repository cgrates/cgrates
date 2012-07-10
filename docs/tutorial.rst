Tutorial
========
The general usage of the CGRateS involves creating a CallDescriptor structure sending it to the balancer via JSON RPC and getting a response from the balancer inf form of a CallCost structure or a numeric value for requested information.

The general steps to get up and running with CGRateS are:

#. Create JSON files containing rates, budgets, tariff plans and destinations, see :ref:`data-importing`.
#. Load the data in the databases using the loader tool.
#. Start the balancer, see :ref:`running`.
#. Start one ore more raters.
#. Make API calls to the balancer/rater.

CallDescriptor structure
------------------------
	- TOR                                int
	- CstmId, Subject, DestinationPrefix string
	- TimeStart, TimeEnd                 time.Time
	- Amount                             float64
TOR
	Type Of Record, used to differentiate between various type of records
CstmId
	Customer Identification used for multi tenant databases
Subject
	Subject for this query
DestinationPrefix
	Destination prefix to be matched
TimeStart, TimeEnd
	The start end end of the call in question
Amount
	The amount requested in various API calls (e.g. DebitSMS amount)

CallCost structure
------------------
	- TOR                                int
	- CstmId, Subject, DestinationPrefix string
	- Cost, ConnectFee                   float64
	- Timespans                          []*TimeSpan
TOR
	Type Of Record, used to differentiate between various type of records (for query identification and confirmation)
CstmId
	Customer Identification used for multi tenant databases (for query identification and confirmation)
Subject
	Subject for this query (for query identification and confirmation)
DestinationPrefix
	Destination prefix to be matched (for query identification and confirmation)
Cost
	The requested cost
ConnectFee
	The requested connection cost
Timespans
	The timespans in witch the initial TimeStart-TimeEnd was split in for cost determination with all pricing and cost information attached. 

Instalation
-----------
**Using packages**
**Using source**

After the go environment is installed_ and setup_ just issue the following commands:
::

	go get github.com/cgrate/cgrates

This will install the sources and compile all available tools	
	
After that navigate

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
  		-httpapiaddr="127.0.0.1:8000": HTTP API server address (localhost:2002)
  		-jsonrpcaddr="127.0.0.1:2001": JSON RPC server address (localhost:2001)
  		-rateraddr="127.0.0.1:2000": Rater server address (localhost:2000)

cgr-rater
	The cgr-rater can be provided with the balancer server address and can be configured to listen to a specific interface and port.
::

	rif@grace:~$ cgr-rater --help
	Usage of cgr-rater:
	  -balancer="127.0.0.1:2000": balancer address host:port
	  -json=false: use json for rpc encoding
	  -listen="127.0.0.1:1234": listening address host:port

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

	rif@grace:~$ cgr-cgrates 
	List of commands:
		getcost
		getmaxsessiontime
		debitbalance
		debitsms
		debitseconds
		addvolumediscountseconds
		resetvolumediscountseconds
		addrecievedcallseconds
		resetuserbudget
		status

cgr-loader
	The loader is the most configurable tool because it has options for each of the three supported databases (kyoto, redis and mongodb).
	Apart from that multi-database options it is quite easy to be used.
	The apfile, destfile, tpfile and ubfile parameters are for specifying the input json files.
	The storage parameter specifies the database to be used and then the databases access information (host:port or file) has to be provided.

	:Example: cgr-loader -storage=kyoto -kyotofile=storage.kch -apfile=activationperiods.json -destfile=destinations.json -tpfile=tariffplans.json -ubfile=userbudgets.json
::

	rif@grace:~$ cgr-loader --help
	Usage of cgr-loader:
	  -apfile="ap.json": Activation Periods containing intervals file
	  -destfile="dest.json": Destinations file
	  -kyotofile="storage.kch": kyoto storage file (storage.kch)
	  -mdb="test": mongo database name (test)
	  -mongoserver="127.0.0.1:27017": mongo server address (127.0.0.1:27017)
	  -pass="": redis database password
	  -rdb=10: redis database number (10)
	  -redisserver="tcp:127.0.0.1:6379": redis server address (tcp:127.0.0.1:6379)
	  -storage="all": kyoto|redis|mongo
	  -tpfile="tp.json": Tariff plans file
	  -ubfile="ub.json": User budgets file

.. _data-importing:

