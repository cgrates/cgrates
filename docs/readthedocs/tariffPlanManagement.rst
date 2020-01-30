6. Tariff Plan Management
=========================

6.1 Create TariffPlan
---------------------

6.2 Assign TariffPlan
---------------------

6.3 Calculate Cost
------------------

 Cost simulator calculates call cost (sum of ConnectFee and Cost fields) for a given pair of source(subject) and destination accounts for a specific time interval. This request can provide Pre Call Cost.

:Hint:
    cgr> cost Tenant="cgrates.org" Category="call" Subject="1003" AnswerTime="2014-08-04T13:00:00Z" Destination="1002" Usage="1m25s"

*Request*

::

    {
    	"method": "APIerSv1.GetCost",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"Category": "call",
    		"Subject": "1003",
    		"AnswerTime": "2014-08-04T13:00:00Z",
    		"Destination": "1002",
    		"Usage": "1m25s"
    	}],
    	"id": 7
    }

*Response*

::

    {
    	"id": 7,
    	"result": {
    		"CGRID": "",
    		"RunID": "",
    		"StartTime": "2014-08-04T13:00:00Z",
    		"Usage": 90000000000,
    		"Cost": 0.25,
    		"Charges": [{
    			"RatingID": "81ca386",
    			"Increments": [{
    				"Usage": 60000000000,
    				"Cost": 0.2,
    				"AccountingID": "",
    				"CompressFactor": 1
    			}],
    			"CompressFactor": 1
    		}, {
    			"RatingID": "2ff21f2",
    			"Increments": [{
    				"Usage": 30000000000,
    				"Cost": 0.05,
    				"AccountingID": "",
    				"CompressFactor": 1
    			}],
    			"CompressFactor": 1
    		}],
    		"AccountSummary": null,
    		"Rating": {
    			"2ff21f2": {
    				"ConnectFee": 0.4,
    				"RoundingMethod": "*up",
    				"RoundingDecimals": 4,
    				"MaxCost": 0,
    				"MaxCostStrategy": "",
    				"TimingID": "998f4c1",
    				"RatesID": "7977f71",
    				"RatingFiltersID": "5165642"
    			},
    			"81ca386": {
    				"ConnectFee": 0.4,
    				"RoundingMethod": "*up",
    				"RoundingDecimals": 4,
    				"MaxCost": 0,
    				"MaxCostStrategy": "",
    				"TimingID": "998f4c1",
    				"RatesID": "e630781",
    				"RatingFiltersID": "5165642"
    			}
    		},
    		"Accounting": {},
    		"RatingFilters": {
    			"5165642": {
    				"DestinationID": "DST_1002",
    				"DestinationPrefix": "1002",
    				"RatingPlanID": "RP_RETAIL2",
    				"Subject": "*out:cgrates.org:call:*any"
    			}
    		},
    		"Rates": {
    			"7977f71": [{
    				"GroupIntervalStart": 0,
    				"Value": 0.2,
    				"RateIncrement": 60000000000,
    				"RateUnit": 60000000000
    			}, {
    				"GroupIntervalStart": 60000000000,
    				"Value": 0.1,
    				"RateIncrement": 30000000000,
    				"RateUnit": 60000000000
    			}],
    			"e630781": [{
    				"GroupIntervalStart": 0,
    				"Value": 0.2,
    				"RateIncrement": 60000000000,
    				"RateUnit": 60000000000
    			}, {
    				"GroupIntervalStart": 60000000000,
    				"Value": 0.1,
    				"RateIncrement": 30000000000,
    				"RateUnit": 60000000000
    			}]
    		},
    		"Timings": {
    			"998f4c1": {
    				"Years": [],
    				"Months": [],
    				"MonthDays": [],
    				"WeekDays": [1, 2, 3, 4, 5],
    				"StartTime": "08:00:00"
    			}
    		}
    	},
    	"error": null
    }

Make Test Call
##############

:Hint:
    initiate test call from account 1003 to 1002
