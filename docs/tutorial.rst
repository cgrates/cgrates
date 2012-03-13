Tutorial
========
The general usage of the CGRateS involves creating a CallDescriptor structure sending it to the balancer via JSON RPC and getting a response from the balancer inf form of a CallCost structure or a numeric value for requested information.

The general steps to get up and running with CGRateS are:

#. Create JSON files containing rates, budgets, tariff plans and destinations, see :ref:`data-importing`.
#. Load the data in the databases using the loader tool.
#. Start the balancer, see :ref:`running`.
#. Start one ore more raters.
#. Make API calls to the balancer.

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

.. image::  images/general.png

Instalation
-----------
**Using packages**
**Using source**

.. _running:

Running
-------

There are only three main command to used with CGRateS:

balancer
	The balancer will open a JSON RPC server and an HTTP server ready for taking external requests. It will also open and rater server on witch the raters will register themselves when they start.
::

	rif@grace:~$ balancer --help
	Usage of balancer:
  		-httpapiaddr="127.0.0.1:8000": HTTP API server address (localhost:2002)
  		-jsonrpcaddr="127.0.0.1:2001": JSON RPC server address (localhost:2001)
  		-rateraddr="127.0.0.1:2000": Rater server address (localhost:2000)

rater
	The rater can be provided with the balancer server address and can be configured to listen to a specific interface and port.
::

	rif@grace:~$ rater --help
	Usage of rater:
	  -listen="127.0.0.1:1234": listening address host:port
	  -balancer="127.0.0.1:2000": balancer address host:port

loader
	The loader is the most configurable tool because it has options for each of the three supported databases (kyoto, redis and mongodb).
	Apart from that multi-database options it is quite easy to be used.
	The apfile, destfile, tpfile and ubfile parameters are for specifying the input json files.
	The storage parameter specifies the database to be used and then the databses access information (host:port or file) has to be provided.

	:Example: loader -storage=kyoto -kyotofile=storage.kch -apfile=activationperiods.json -destfile=destinations.json -tpfile=tariffplans.json -ubfile=userbudgets.json
::

	rif@grace:~$ loader --help
	Usage of loader:
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

Data importing
--------------

**Activation periods**
::
	{"TOR": 0,"CstmId":"vdf","Subject":"rif","DestinationPrefix":"0257", "ActivationPeriods": [
	        {"ActivationTime": "2012-01-01T00:00:00Z", "Intervals": [
	                {"BillingUnit":1,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":0.1,
	                	"StartTime":"18:00:00","EndTime":"","WeekDays":[1,2,3,4,5]},
	                {"BillingUnit":1,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":0.2,
	                	"StartTime":"","EndTime":"18:00:00","WeekDays":[1,2,3,4,5]}, 
	                {"BillingUnit":1,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":0.1,
	                	"StartTime":"","EndTime":"","WeekDays":[6,0]}
	            ]
	        },
	        {"ActivationTime": "2012-02-08T00:00:00Z", "Intervals": [                
	                {"BillingUnit":60,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":10,
	                	"StartTime":"","EndTime":"18:00:00","WeekDays":[1,2,3,4,5]}, 
	                {"BillingUnit":60,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":1,
	                	"StartTime":"18:00:00","EndTime":"","WeekDays":[1,2,3,4,5]},
	                {"BillingUnit":60,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":1,
	                	"StartTime":"","EndTime":"","WeekDays":[6,0]}
	            ]
	        }
	    ]     
	}

**Destinations**
::
	{"Id":"nationale", "Prefixes":["0256","0257","0723","0740"]},
	{"Id":"retea", "Prefixes":["0723","0724"]},
	{"Id":"mobil", "Prefixes":["0723","0740"]},
	{"Id":"radu", "Prefixes":["0723045326"]}


**Tariff plans**
::
	{"Id":"dimineata","SmsCredit":100,"ReceivedCallsSecondsLimit": 100,
			"RecivedCallBonus" : {"Credit": 100},
			"MinuteBuckets":
				[{"Seconds":100,"Priority":10,"Price":0.01,"DestinationId":"nationale"},
					{"Seconds":1000,"Priority":20,"Price":0,"DestinationId":"retea"}],
			"VolumeDiscountThresholds":
				[{"Volume": 100, "Discount": 10},
					{"Volume": 500, "Discount": 15},
					{"Volume": 1000, "Discount": 20}]			
	}

**User budgets**
::
	{"Id":"broker","Credit":0,"SmsCredit":0,"Traffic":0,"VolumeDiscountSeconds":0,
		"ReceivedCallSeconds":0,"ResetDayOfTheMonth":10,"TariffPlanId":"seara","MinuteBuckets":
	    	[{"Seconds":10,"Priority":10,"Price":0.01,"DestinationId":"nationale"},
		 		{"Seconds":100,"Priority":20,"Price":0,"DestinationId":"retea"}]
	}

Database selection
-------------------

**Kyoto cabinet**

Pros:
	- super fast (the in memory data is accessed directly by the rater processes)
	- easy backup
Cons:
	- harder to synchronize different raters	

**Redis**

Pros:
	- easy configuration
	- easy master-server configuration	
Cons:
	- slower than kyoto
	- less features than mongodb

**MongoDB**

Pros:
	- most features
	- most advanced clustering options
Cons:
	- slowest of the three
