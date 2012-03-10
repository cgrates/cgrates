Tutorial
========
The general usage of the cgrates involves creating a CallDescriptor stucture sending it to the balancer via JSON RPC and getting a response from the balancer inf form of a CallCost structure or a numeric value for requested information.

CallDescriptor structure
------------------------
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

CallCost structure
------------------
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
**Using packages**
**Using source**
Running
-------

Data importing
--------------

**Activation periods**


{"TOR": 0,"CstmId":"vdf","Subject":"rif","DestinationPrefix":"0257", "ActivationPeriods": [
        {"ActivationTime": "2012-01-01T00:00:00Z", "Intervals": [
                {"BillingUnit":1,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":0.1,"StartTime":"18:00:00","EndTime":"","WeekDays":[1,2,3,4,5]},
                {"BillingUnit":1,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":0.2,"StartTime":"","EndTime":"18:00:00","WeekDays":[1,2,3,4,5]}, 
                {"BillingUnit":1,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":0.1,"StartTime":"","EndTime":"","WeekDays":[6,0]}
            ]
        },
        {"ActivationTime": "2012-02-08T00:00:00Z", "Intervals": [                
                {"BillingUnit":60,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":10,"StartTime":"","EndTime":"18:00:00","WeekDays":[1,2,3,4,5]}, 
                {"BillingUnit":60,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":1,"StartTime":"18:00:00","EndTime":"","WeekDays":[1,2,3,4,5]},
                {"BillingUnit":60,"ConnectFee":0,"Month":0,"MonthDay":0,"Ponder":0,"Price":1,"StartTime":"","EndTime":"","WeekDays":[6,0]}
            ]
        }
    ]     
},


**Destinations**

{"Id":"nationale", "Prefixes":["0256","0257","0723","0740"]},
{"Id":"retea", "Prefixes":["0723","0724"]},
{"Id":"mobil", "Prefixes":["0723","0740"]},
{"Id":"radu", "Prefixes":["0723045326"]}


**Tariff plans**

{"Id":"dimineata","SmsCredit":100,"ReceivedCallsSecondsLimit": 100,
		"RecivedCallBonus" : {"Credit": 100},
		"MinuteBuckets":
			[{"Seconds":100,"Priority":10,"Price":0.01,"DestinationId":"nationale"}, {"Seconds":1000,"Priority":20,"Price":0,"DestinationId":"retea"}],
		"VolumeDiscountThresholds":
			[{"Volume": 100, "Discount": 10},{"Volume": 500, "Discount": 15},{"Volume": 1000, "Discount": 20}]			
}

**User budgets**

{"Id":"broker","Credit":0,"SmsCredit":0,"Traffic":0,"VolumeDiscountSeconds":0,"ReceivedCallSeconds":0,"ResetDayOfTheMonth":10,"TariffPlanId":"seara","MinuteBuckets":
    [{"Seconds":10,"Priority":10,"Price":0.01,"DestinationId":"nationale"},
	 {"Seconds":100,"Priority":20,"Price":0,"DestinationId":"retea"}]}