API Calls
=========
The general API usage of the CGRateS involves creating a CallDescriptor structure sending it to the balancer via JSON/GOB RPC and getting a response from the balancer in form of a CallCost structure or a numeric value for requested information.

CallDescriptor structure
------------------------
	- Direction, TOR, Tenant, Subject, Account, DestinationPrefix string
	- TimeStart, TimeEnd                 Time
	- Amount                             float64

Direction
	The direction of the call (inbound or outbound)
TOR
	Type Of Record, used to differentiate between various type of records
Tenant
	Customer Identification used for multi tenant databases
Subject
	Subject for this query
Account
	Used when different from subject
Destination
	Destination call id to be matched
TimeStart, TimeEnd
	The start end end of the call in question
Amount
	The amount requested in various API calls (e.g. DebitSMS amount)

The **Subject** field is used usually used to identify both the client in the detailed cost list and the user in the balances database. When there is some additional info added to the subject for the call price list then the **Account** attribute is used to specify the balance for the client. For example: the subject can be rif:from:ha or rif:form:mu and for both we would use the rif account.


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

As stated before the balancer (or the rater directly) can be accesed via json rpc. 

The smallest python snippet to acces the CGRateS balancer is this:

::

	cd = {"Tor":0,
		"CstmId": "vdf",
		"Subject": "rif",
		"DestinationPrefix": "0256",
		"TimeStart": "2012-02-02T17:30:00Z",
		"TimeEnd": "2012-02-02T18:30:00Z"}

	s = socket.create_connection(("127.0.0.1", 2001))
	s.sendall(json.dumps(({"id": 1, "method": "Responder.Get", "params": [cd]})))
	print s.recv(4096)

This also gives you a pretty good idea of how JSON-RPC works. You can find details in the specification_. A call to a JSON-RPC server simply sends a block of data through a socket. The data is formatted as a JSON structure, and a call consists of an id (so you can sort out the results when they come back), the name of the method to execute on the server, and params, an array of parameters which can itself consist of complex JSON objects. The dumps() call converts the Python structure into JSON.

.. _specification:  http://json-rpc.org/wiki/specification

In the stress folder you can find a better example of python client using a class that reduces the actual call code to::

	rpc =JSONClient(("127.0.0.1", 2001))
	result = rpc.call("Responder.Get", cd)
	print result
	
Call API
--------
GetCost
	Creates a CallCost structure with the cost information calculated for the received CallDescriptor.

Debit
    Interface method used to add/substract an amount of cents or bonus seconds (as returned by GetCost method) from user's money balance.


MaxDebit
    Interface method used to add/substract an amount of cents or bonus seconds (as returned by GetCost method) from user's money balance.
    This methods combines the Debit and GetMaxSessionTime and will debit the max available time as returned by the GetMaxSessionTime method. The amount filed has to be filled in call descriptor.


DebitBalance
	Interface method used to add/substract an amount of cents from user's money budget.
	The amount filed has to be filled in call descriptor.

DebitSMS
	Interface method used to add/substract an amount of units from user's SMS budget.
	The amount filed has to be filled in call descriptor.

DebitSeconds
	Interface method used to add/substract an amount of seconds from user's minutes budget.
	The amount filed has to be filled in call descriptor.

GetMaxSessionTime
	Returns the approximate max allowed session for user budget. It will try the max amount received in the call descriptor 
	and will decrease it by 10% for nine times. So if the user has little credit it will still allow 10% of the initial amount.
	If the user has no credit then it will return 0.

AddRecievedCallSeconds
	Adds the specified amount of seconds to the received call seconds. When the threshold specified in the user's tariff plan is reached then the received call budget is reseted and the bonus specified in the tariff plan is applied.
	The amount filed has to be filled in call descriptor.

FlushCache
    Cleans all internal cached (Destinations, RatingProfiles)


Tariff plan importer APIs
-------------------------

These operate on a tpid and are used to import the tariff plan content into storDb.

TariffPlan
~~~~~~~~~~

.. toctree::
   :maxdepth: 2

   api_tp

Timings
~~~~~~~

.. toctree::
   :maxdepth: 2

   api_tptimings

Destinations
~~~~~~~~~~~~

.. toctree::
   :maxdepth: 2

   api_tpdestinations

Rates
~~~~~

.. toctree::
   :maxdepth: 2

   api_tprates

DestinationRates
~~~~~~~~~~~~~~~~

.. toctree::
   :maxdepth: 2

   api_tpdestinationrates

DestinationRateTimings
~~~~~~~~~~~~~~~~~~~~~~

.. toctree::
   :maxdepth: 2

   api_tpdestratetimings

RatingProfiles
~~~~~~~~~~~~~~

.. toctree::
   :maxdepth: 2

   api_tpratingprofiles

Actions
~~~~~~~

.. toctree::
   :maxdepth: 2

   api_tpactions

ActionTimings
~~~~~~~~~~~~~


.. toctree::
   :maxdepth: 2

   api_tpactiontimings

ActionTriggers
~~~~~~~~~~~~~~


.. toctree::
   :maxdepth: 2

   api_tpactiontriggers

AccountActions
~~~~~~~~~~~~~~

.. toctree::
   :maxdepth: 2

   api_tpaccountactions


Management API
--------------
These operate on live data.

GetDestination
++++++++++++++

Gets the destinations for the specified tag.

::

	type AttrDestination struct {
		Id       string
		Prefixes []string
	}

Only the Id field must be set for get operation.

Example
	GetDestination(attr \*AttrDestination, reply \*AttrDestination)


GetBalance
++++++++++

Gets a specific balance of a user acoount.

::

   type AttrGetBalance struct {
		Tenant    string
		Account   string
		BalanceId string
		Direction string
	}

The Tenant is the network tenant of the account.

The Account is the id of the account for which the balance is desired.

The BalanceId can have one of the following string values: MONETARY, SMS, INTERNET, INTERNET_TIME, MINUTES.

Direction can be the strings IN or OUT (default OUT).

Return value is the balance value as float64.

Example
    GetBalance(attr \*AttrGetBalance, reply \*float64)

AddBalance
++++++++++

Adds an amount to a specific balance of a user account.

::

    type AttrAddBalance struct {
		Tenant    string
		Account   string
		BalanceId string
		Direction string
		Value     float64
	}

The Tenant is the network tenant of the account.

The Account is the id of the account for which the balance is set.

The BalanceId can have one of the following string values: MONETARY, SMS, INTERNET, INTERNET_TIME, MINUTES.

Direction can be the strings IN or OUT (default OUT).

Value is the amount to be added to the specified balance.

Example
     AddBalance(attr \*AttrAddBalance, reply \*float64)


ExecuteAction
+++++++++++++

Executes specified action on a user account.

::

    type AttrExecuteAction struct {
		Direction string
		Tenant    string
		Account   string
		ActionsId string
	}

Example
    ExecuteAction(attr \*AttrExecuteAction, reply \*float64)

AddTriggeredAction
++++++++++++++++++

::

	type AttrAddActionTrigger struct {
		Tenant         string
		Account        string
		Direction      string
		BalanceId      string
		ThresholdValue float64
		DestinationId  string
		Weight         float64
		ActionsId      string
	}

Example
	AddTriggeredAction(attr \*AttrAddActionTrigger, reply \*float64)

AddAcount
+++++++++

::

	type AttrAddAccount struct {
		Tenant          string
		Direction       string
		Account         string
		Type            string // prepaid-postpaid
		ActionTimingsId string
	}
	
Example
	AddAccount(attr \*AttrAddAccount, reply \*float64)




RatingProfiles
~~~~~~~~~~~~~~

.. toctree::
   :maxdepth: 2

   api_ratingprofiles

Accounts
~~~~~~~~

.. toctree::
   :maxdepth: 2

   api_accounts


Administration APIs
-------------------

.. toctree::
   :maxdepth: 2

   api_cdrs
   api_cache
   api_scheduler



