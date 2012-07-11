Api Calls
========
The general API usage of the CGRateS involves creating a CallDescriptor structure sending it to the balancer via JSON/GOB RPC and getting a response from the balancer in form of a CallCost structure or a numeric value for requested information.

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

In the stress folder you can find a better example of python client using a class that reduces tha ctual call code to::

	rpc =JSONClient(("127.0.0.1", 2001))
	result = rpc.call("Responder.Get", cd)
	print result
	
JSON RPC
--------
GetCost
	Creates a CallCost structure with the cost information calculated for the received CallDescriptor.

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

AddVolumeDiscountSeconds
	Interface method used to add an amount to the accumulated placed call seconds to be used for volume discount.
	The amount filed has to be filled in call descriptor.

ResetVolumeDiscountSeconds
	Resets the accumulated volume discount seconds (to zero).

AddRecievedCallSeconds
	Adds the specified amount of seconds to the received call seconds. When the threshold specified in the user's tariff plan is reached then the received call budget is reseted and the bonus specified in the tariff plan is applied.
	The amount filed has to be filled in call descriptor.

ResetUserBudget
	Resets user budgets value to the amounts specified in the tariff plan.

HTTP
----

getcost
	:Example: curl "http://127.0.0.1:8000/getcost?cstmid=vdf&subj=rif&dest=0257"
debitbalance
	:Example: curl "http://127.0.0.1:8000/debitbalance?cstmid=vdf&subj=rif&dest=0257@amount=100"
debitsms
	:Example: curl "http://127.0.0.1:8000/debitsms?cstmid=vdf&subj=rif&dest=0257@amount=100"
debitseconds
	:Example: curl "http://127.0.0.1:8000/debitseconds?cstmid=vdf&subj=rif&dest=0257@amount=100"
getmaxsessiontime
	:Example: curl "http://127.0.0.1:8000/getmaxsessiontime?cstmid=vdf&subj=rif&dest=0257@amount=100"
addvolumediscountseconds
	:Example: curl "http://127.0.0.1:8000/addvolumediscountseconds?cstmid=vdf&subj=rif&dest=0257@amount=100"
resetvolumediscountseconds
	:Example: curl "http://127.0.0.1:8000/resetvolumediscountseconds?cstmid=vdf&subj=rif&dest=0257"
addrecievedcallseconds
	:Example: curl "http://127.0.0.1:8000/addrecievedcallseconds?cstmid=vdf&subj=rif&dest=0257@amount=100"
resetuserbudget
	:Example: curl "http://127.0.0.1:8000/resetuserbudget?cstmid=vdf&subj=rif&dest=0257"
