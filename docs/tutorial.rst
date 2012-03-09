Tutorial
========
The general usage of the cgrates involves creating a CallDescriptor stucture sending it to the balancer via JSON RPC and getting a response from the balancer inf form of a CallCost structure or a numeric value for requested information.

CallDescriptor struct
	- TOR                                int
	- CstmId, Subject, DestinationPrefix string
	- TimeStart, TimeEnd                 time.Time
	- Amount                             float64
TOR
	Type Of Record, used to differentiate between various type of records
CstmId
	Customer Identification used for multi tennant databases
Subject
	Subject for this query
DestinationPrefix
	Destination prefix to be matched
TimeStart, TimeEnd
	The start end end of the call in question
Amount
	The amount requested in various api calss (e.g. DebitSMS amount)

CallCost struct 
	- TOR                                int
	- CstmId, Subject, DestinationPrefix string
	- Cost, ConnectFee                   float64
	- Timespans                          []*TimeSpan
TOR
	Type Of Record, used to differentiate between various type of records (for query identification and confirmation)
CstmId
	Customer Identification used for multi tennant databases (for query identification and confirmation)
Subject
	Subject for this query (for query identification and confirmation)
DestinationPrefix
	Destination prefix to be matched (for query identification and confirmation)
Cost
	The requested cost
ConnectFee
	The requested connection cost
Timespans
	The timespans in wicht the initial TimeStart-TimeEnd was split in for cost determination with all pricingg and cost information attached. 

.. image::  images/general.png

Instalation
-----------

Running
-------

Data importing
--------------
