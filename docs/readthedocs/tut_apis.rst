.. CGRateS_JSON_APIs documentation master file, created by
   sphinx-quickstart on Tue Dec  5 13:24:02 2017.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

CGRateS API Document
====================

CGRateS billing solution allow users provisioning and tarif plan management.

*Usage example through postman*

URL: http://192.168.10.17:2080/jsonrpc
::

    {"method":"ApierV2.GetAccounts","params":[{"Tenant":"cgrates.org","AccountIds":["1001"],"Offset":0,"Limit":0}],"id":3}
    Content-Type: application/json

:Hint:
    below cgr-console prompt as 'cgr>'

User Life Cycle
===============

Following steps will cover use-case "User Life cycle"

User Management
---------------

Create user account
###################

:Hint:
    cgr> account_set Tenant="cgrates.org" Account="1003" ActionPlanIDs=["PACKAGE_10"] ActionTriggerIDs=["STANDARD_TRIGGERS"]

*Request*

::

    {
    	"method": "ApierV2.SetAccount",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"Account": "1003",
    		"ActionPlanIDs": ["PACKAGE_10"],
    		"ActionPlansOverwrite": false,
    		"ActionTriggerIDs": ["STANDARD_TRIGGERS"],
    		"ActionTriggerOverwrite": false,
    		"AllowNegative": null,
    		"Disabled": null,
    		"ReloadScheduler": false
    	}],
    	"id": 0
    }

*Response*

::

    {"id": 0,"result": "OK","error": null}

Get user account
################

:Hint:
    cgr> accounts Tenant="cgrates.org" AccountIds=["1003"]

*Request*

::

    {
    	"method": "ApierV2.GetAccounts",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"AccountIds": ["1003"],
    		"Offset": 0,
    		"Limit": 0
    	}],
    	"id": 1
    }

*Response*

::

    {
    	"id": 1,
    	"result": [{
    		"ID": "cgrates.org:1003",
    		"BalanceMap": {
    			"*monetary": [{
    				"Uuid": "df24bcbd-d0e2-4a67-a188-7c9621ae81d7",
    				"ID": "",
    				"Value": 0.15,
    				"Directions": {
    					"*out": true
    				},
    				"ExpirationDate": "0001-01-01T00:00:00Z",
    				"Weight": 10,
    				"DestinationIDs": {},
    				"RatingSubject": "",
    				"Categories": {},
    				"SharedGroups": {},
    				"Timings": [],
    				"TimingIDs": {},
    				"Disabled": false,
    				"Factor": {},
    				"Blocker": false
    			}, {
    				"Uuid": "9a22f090-a49a-4da6-bdca-5e40810e4b18",
    				"ID": "23456",
    				"Value": 12,
    				"Directions": {},
    				"ExpirationDate": "0001-01-01T00:00:00Z",
    				"Weight": 0,
    				"DestinationIDs": {},
    				"RatingSubject": "",
    				"Categories": {},
    				"SharedGroups": {},
    				"Timings": [],
    				"TimingIDs": {},
    				"Disabled": false,
    				"Factor": {},
    				"Blocker": false
    			}, {
    				"Uuid": "4da21ba2-d899-49b1-ae60-3e8a237a49bb",
    				"ID": "123456",
    				"Value": 0.2,
    				"Directions": {
    					"*out": true
    				},
    				"ExpirationDate": "0001-01-01T00:00:00Z",
    				"Weight": 0,
    				"DestinationIDs": {},
    				"RatingSubject": "",
    				"Categories": {},
    				"SharedGroups": {},
    				"Timings": [],
    				"TimingIDs": {},
    				"Disabled": false,
    				"Factor": {},
    				"Blocker": false
    			}]
    		},
    		"UnitCounters": {
    			"*monetary": [{
    				"CounterType": "*event",
    				"Counters": [{
    					"Value": 0,
    					"Filter": {
    						"Uuid": null,
    						"ID": "df4d286a-445f-40a8-ab84-215153d4f2ac",
    						"Type": "*monetary",
    						"Value": null,
    						"Directions": {
    							"*out": true
    						},
    						"ExpirationDate": null,
    						"Weight": null,
    						"DestinationIDs": {
    							"FS_USERS": true
    						},
    						"RatingSubject": null,
    						"Categories": null,
    						"SharedGroups": null,
    						"TimingIDs": null,
    						"Timings": [],
    						"Disabled": null,
    						"Factor": null,
    						"Blocker": null
    					}
    				}]
    			}]
    		},
    		"ActionTriggers": [{
    			"ID": "STANDARD_TRIGGERS",
    			"UniqueID": "621cb77f-c427-445f-8dfc-05b8105a1709",
    			"ThresholdType": "*min_balance",
    			"ThresholdValue": 2,
    			"Recurrent": false,
    			"MinSleep": 0,
    			"ExpirationDate": "0001-01-01T00:00:00Z",
    			"ActivationDate": "0001-01-01T00:00:00Z",
    			"Balance": {
    				"Uuid": null,
    				"ID": null,
    				"Type": "*monetary",
    				"Value": null,
    				"Directions": {
    					"*out": true
    				},
    				"ExpirationDate": null,
    				"Weight": null,
    				"DestinationIDs": null,
    				"RatingSubject": null,
    				"Categories": null,
    				"SharedGroups": null,
    				"TimingIDs": null,
    				"Timings": [],
    				"Disabled": null,
    				"Factor": null,
    				"Blocker": null
    			},
    			"Weight": 10,
    			"ActionsID": "LOG_WARNING",
    			"MinQueuedItems": 0,
    			"Executed": true,
    			"LastExecutionTime": "2017-12-12T15:19:45.742Z"
    		}, {
    			"ID": "STANDARD_TRIGGERS",
    			"UniqueID": "df4d286a-445f-40a8-ab84-215153d4f2ac",
    			"ThresholdType": "*max_event_counter",
    			"ThresholdValue": 5,
    			"Recurrent": false,
    			"MinSleep": 0,
    			"ExpirationDate": "0001-01-01T00:00:00Z",
    			"ActivationDate": "0001-01-01T00:00:00Z",
    			"Balance": {
    				"Uuid": null,
    				"ID": "df4d286a-445f-40a8-ab84-215153d4f2ac",
    				"Type": "*monetary",
    				"Value": null,
    				"Directions": {
    					"*out": true
    				},
    				"ExpirationDate": null,
    				"Weight": null,
    				"DestinationIDs": {
    					"FS_USERS": true
    				},
    				"RatingSubject": null,
    				"Categories": null,
    				"SharedGroups": null,
    				"TimingIDs": null,
    				"Timings": [],
    				"Disabled": null,
    				"Factor": null,
    				"Blocker": null
    			},
    			"Weight": 10,
    			"ActionsID": "LOG_WARNING",
    			"MinQueuedItems": 0,
    			"Executed": false,
    			"LastExecutionTime": "0001-01-01T00:00:00Z"
    		}, {
    			"ID": "STANDARD_TRIGGERS",
    			"UniqueID": "cb60f788-6077-4f3c-b8b2-4d1ba3077abc",
    			"ThresholdType": "*max_balance",
    			"ThresholdValue": 20,
    			"Recurrent": false,
    			"MinSleep": 0,
    			"ExpirationDate": "0001-01-01T00:00:00Z",
    			"ActivationDate": "0001-01-01T00:00:00Z",
    			"Balance": {
    				"Uuid": null,
    				"ID": null,
    				"Type": "*monetary",
    				"Value": null,
    				"Directions": {
    					"*out": true
    				},
    				"ExpirationDate": null,
    				"Weight": null,
    				"DestinationIDs": null,
    				"RatingSubject": null,
    				"Categories": null,
    				"SharedGroups": null,
    				"TimingIDs": null,
    				"Timings": [],
    				"Disabled": null,
    				"Factor": null,
    				"Blocker": null
    			},
    			"Weight": 10,
    			"ActionsID": "LOG_WARNING",
    			"MinQueuedItems": 0,
    			"Executed": false,
    			"LastExecutionTime": "0001-01-01T00:00:00Z"
    		}, {
    			"ID": "STANDARD_TRIGGERS",
    			"UniqueID": "7f7621f4-6074-4502-bbc0-a8aeca7c1008",
    			"ThresholdType": "*max_balance",
    			"ThresholdValue": 100,
    			"Recurrent": false,
    			"MinSleep": 0,
    			"ExpirationDate": "0001-01-01T00:00:00Z",
    			"ActivationDate": "0001-01-01T00:00:00Z",
    			"Balance": {
    				"Uuid": null,
    				"ID": null,
    				"Type": "*monetary",
    				"Value": null,
    				"Directions": {
    					"*out": true
    				},
    				"ExpirationDate": null,
    				"Weight": null,
    				"DestinationIDs": null,
    				"RatingSubject": null,
    				"Categories": null,
    				"SharedGroups": null,
    				"TimingIDs": null,
    				"Timings": [],
    				"Disabled": null,
    				"Factor": null,
    				"Blocker": null
    			},
    			"Weight": 10,
    			"ActionsID": "DISABLE_AND_LOG",
    			"MinQueuedItems": 0,
    			"Executed": false,
    			"LastExecutionTime": "0001-01-01T00:00:00Z"
    		}],
    		"AllowNegative": false,
    		"Disabled": false
    	}],
    	"error": null
    }

Remove user account
###################

:Hint:
    cgr> account_remove Tenant="cgrates.org" Account="1003"

*Request*

::

    {
    	"method": "ApierV1.RemoveAccount",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"Account": "1003",
    		"ReloadScheduler": false
    	}],
    	"id": 3
    }

*Response*

::

    {"id": 3,"result": "OK","error": null}

Balance Management
------------------

Balance set
###########

 replaces existing value of BalanceID '23456' with value 12 for account 1003 belongs to tenant 'cgrates.org'

:Hint:
    cgr> balance_set Tenant="cgrates.org" Account="1003" Direction="\*out" Value=12 BalanceID="23456"

*Request*

::

    {
    	"method": "ApierV1.SetBalance",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"Account": "1003",
    		"BalanceType": "*monetary",
    		"BalanceUUID": null,
    		"BalanceID": "23456",
    		"Directions": null,
    		"Value": 12,
    		"ExpiryTime": null,
    		"RatingSubject": null,
    		"Categories": null,
    		"DestinationIds": null,
    		"TimingIds": null,
    		"Weight": null,
    		"SharedGroups": null,
    		"Blocker": null,
    		"Disabled": null
    	}],
    	"id": 6
    }


*Response*

::

    {"id":6,"result":"OK","error":null}

Balance add
###########

 adds 10 cent to account=1003 where tenant=cgrates.org

:Hint:
    cgr> balance_add Tenant="cgrates.org" Account="1003" BalanceId="123456" Value=10

*Request*

::

    {
    	"method": "ApierV1.AddBalance",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"Account": "1003",
    		"BalanceUuid": null,
    		"BalanceId": "123456",
    		"BalanceType": "*monetary",
    		"Directions": null,
    		"Value": 10,
    		"ExpiryTime": null,
    		"RatingSubject": null,
    		"Categories": null,
    		"DestinationIds": null,
    		"TimingIds": null,
    		"Weight": null,
    		"SharedGroups": null,
    		"Overwrite": false,
    		"Blocker": null,
    		"Disabled": null
    	}],
    	"id": 4
    }

*Response*

::

    {"id":4,"result":"OK","error":null}


Balance debit
#############

 deducts 5 cents from account 1003 of tenant cgrates.org

:Hint:
    cgr> balance_debit Tenant="cgrates.org" Account="1003" BalanceId="23456" Value=5 BalanceType="\*monetary"

*Request*

::

    {
    	"method": "ApierV1.DebitBalance",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"Account": "1003",
    		"BalanceUuid": null,
    		"BalanceId": "23456",
    		"BalanceType": "*monetary",
    		"Directions": null,
    		"Value": 5,
    		"ExpiryTime": null,
    		"RatingSubject": null,
    		"Categories": null,
    		"DestinationIds": null,
    		"TimingIds": null,
    		"Weight": null,
    		"SharedGroups": null,
    		"Overwrite": false,
    		"Blocker": null,
    		"Disabled": null
    	}],
    	"id": 5
    }

*Response*

::

    {"id":5,"result":"OK","error":null}


Get Remaining Balance
#####################

Sum of BalanceMap.Value resulted from ApierV2.GetAccounts request


Tariff Plan Management
----------------------

#Create TariffPlan
#Assign TariffPlan

Calculate Cost
######################

 calculates call cost (sum of ConnectFee and Cost fields) for a given pair or source and destination accounts for a specific time interval. This request can provide Pre Call Cost.

:Hint:
    cgr> cost Tenant="cgrates.org" Category="call" Subject="1003" AnswerTime="2014-08-04T13:00:00Z" Destination="1002" Usage="1m25s"

*Request*

::

    {
    	"method": "ApierV1.GetCost",
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

CDR Management
--------------

Export CDRs
###########

:Hint:
    cgr > cdrs_export CdrFormat="csv" ExportDir="/tmp"

*Request*

::

    {
    	"method": "ApierV1.ExportCDRs",
    	"params": [{
    		"ExportTemplate": null,
    		"ExportFormat": null,
    		"ExportPath": null,
    		"Synchronous": null,
    		"Attempts": null,
    		"FieldSeparator": null,
    		"UsageMultiplyFactor": null,
    		"CostMultiplyFactor": null,
    		"ExportID": null,
    		"ExportFileName": null,
    		"RoundingDecimals": null,
    		"Verbose": false,
    		"CGRIDs": null,
    		"NotCGRIDs": null,
    		"RunIDs": null,
    		"NotRunIDs": null,
    		"OriginHosts": null,
    		"NotOriginHosts": null,
    		"Sources": null,
    		"NotSources": null,
    		"ToRs": null,
    		"NotToRs": null,
    		"RequestTypes": null,
    		"NotRequestTypes": null,
    		"Tenants": null,
    		"NotTenants": null,
    		"Categories": null,
    		"NotCategories": null,
    		"Accounts": null,
    		"NotAccounts": null,
    		"Subjects": null,
    		"NotSubjects": null,
    		"DestinationPrefixes": null,
    		"NotDestinationPrefixes": null,
    		"Costs": null,
    		"NotCosts": null,
    		"ExtraFields": null,
    		"NotExtraFields": null,
    		"OrderIDStart": null,
    		"OrderIDEnd": null,
    		"SetupTimeStart": "",
    		"SetupTimeEnd": "",
    		"AnswerTimeStart": "",
    		"AnswerTimeEnd": "",
    		"CreatedAtStart": "",
    		"CreatedAtEnd": "",
    		"UpdatedAtStart": "",
    		"UpdatedAtEnd": "",
    		"MinUsage": "",
    		"MaxUsage": "",
    		"MinCost": null,
    		"MaxCost": null,
    		"Limit": null,
    		"Offset": null,
    		"SearchTerm": ""
    	}],
    	"id": 8
    }

*Response*

::

    {
    	"id": 8,
    	"result": {
    		"ExportedPath": "/var/spool/cgrates/cdre/cdre_1513199075.csv",
    		"TotalRecords": 186,
    		"TotalCost": 56.4371,
    		"FirstOrderID": 1513066080275428946,
    		"LastOrderID": 1513066080275429038,
    		"ExportedCGRIDs": null,
    		"UnexportedCGRIDs": null
    	},
    	"error": null
    }


Or fetch CDRs from mongodb

List all CDRs
#############

:Hint:
    db.getCollection('cdrs').find({})

Filter based on 'cgrid'
#######################

:Hint:
    db.cdrs.find({"cgrid":"84bde1fd133f70572e05e699ea2f1de201e18269", "runid":"\*default"})

Filter calls from 1001 to 1002
##############################

:Hint:
    db.cdrs.find({"account":"1001", "destination":"1002"})

Filter calls from 1003 to 1002
##############################

:Hint:
    db.cdrs.find({"account":"1003", "destination":"1002"})

Filter calls on setup time
##########################

:Hint:

    db.cdrs.find({"setuptime" : ISODate("2017-12-11T23:38:57.000Z")})

LCR Strategy: (\*static)
########################

Use supplier base on LCR rules

:Hint:
    cgr> lcr Account="1001" Destination="1002"

LCR Strategy: (\*lowest_cost)
#############################

Use supplier with least cost

:Hint:
    cgr> lcr Account="1005" Destination="1001"

LCR Strategy: (\*highest_cost)
##############################

Use supplier with highest cost

:Hint:
    cgr> lcr Account="1002" Destination="1002"

LCR Strategy: (\*qos_threshold)
###############################

Use supplier with lowest cost, matching QoS thresholds min/max ASR, ACD, TCD, ACC, TCC

:Hint:
    cgr> lcr Account="1002" Destination="1002"

LCR Strategy: (\*qos)
#####################

Use supplier with best quality, independent of cost

:Hint:
    cgr> lcr Account="1002" Destination="1005"


-------------------------------------------------------------------

GetCacheStats
#############

GetCacheStats returns datadb cache status. Empty params return all stats:

:Hint:

    cgr> cache_stats

*Request*

::

   {
   	"method": "ApierV1.GetCacheStats",
   	"params": [{}],
   	"id": 0
   }

*Response:*

::

   {
   	"id": 0,
   	"result": {
   		"Destinations": 0,
   		"ReverseDestinations": 0,
   		"RatingPlans": 4,
   		"RatingProfiles": 0,
   		"Actions": 0,
   		"ActionPlans": 4,
   		"AccountActionPlans": 0,
   		"SharedGroups": 0,
   		"DerivedChargers": 0,
   		"LcrProfiles": 0,
   		"CdrStats": 6,
   		"Users": 3,
   		"Aliases": 0,
   		"ReverseAliases": 0,
   		"ResourceProfiles": 0,
   		"Resources": 0,
   		"StatQueues": 0,
   		"StatQueueProfiles": 0,
   		"Thresholds": 0,
   		"ThresholdProfiles": 0,
   		"Filters": 0
   	},
   	"error": null
   }

Get Users Profile
#################

GetUsers returns list of all users profile:

:Hint:
    cgr> users

*Request*

::

   {
   	"method": "UsersV1.GetUsers",
   	"params": [{
   		"Tenant": "",
   		"UserName": "",
   		"Masked": false,
   		"Profile": null,
   		"Weight": 0
   	}],
   	"id": 2
   }

*Response*

::

   {
   	"id": 2,
   	"result": [{
   			"Tenant": "cgrates.org",
   			"UserName": "1001",
   			"Masked": false,
   			"Profile": {
   				"Account": "1001",
   				"Cli": "+4986517174963",
   				"RequestType": "*prepaid",
   				"Subject": "1001",
   				"SubscriberId": "1001",
   				"SysPassword": "hisPass321",
   				"SysUserName": "danb",
   				"Uuid": "388539dfd4f5cefee8f488b78c6c244b9e19138e"
   			},
   			"Weight": 0
   		},

   		{
   			"Tenant": "cgrates.org",
   			"UserName": "1002",
   			"Masked": false,
   			"Profile": {
   				"Account": "1002",
   				"RifAttr": "RifVal",
   				"Subject": "1002",
   				"SubscriberId": "1002",
   				"SysUserName": "rif",
   				"Uuid": "27f37edec0670fa34cf79076b80ef5021e39c5b5"
   			},
   			"Weight": 0
   		},

   		{
   			"Tenant": "cgrates.org",
   			"UserName": "1004",
   			"Masked": false,
   			"Profile": {
   				"Account": "1004",
   				"Cli": "+4986517174964",
   				"RequestType": "*rated",
   				"Subject": "1004",
   				"SubscriberId": "1004",
   				"SysPassword": "hisPass321",
   				"SysUserName": "danb4"
   			},
   			"Weight": 0
   		}
   	],
   	"error": null
   }

Get Profile UserName 1001
#########################

Returns a User Profile of user account 1001:

:Hint:

   cgr> users UserName="1001"

*Request*

::

    {
    	"method": "UsersV1.GetUsers",
    	"params": [{
    		"Tenant": "",
    		"UserName": "1001",
    		"Masked": false,
    		"Profile": null,
    		"Weight": 0
    	}],
    	"id": 2
    }

*Response*

::

    {
    	"id": 2,
    	"result": [{
    		"Tenant": "cgrates.org",
    		"UserName": "1001",
    		"Masked": false,
    		"Profile": {
    			"Account": "1001",
    			"Cli": "+4986517174963",
    			"RequestType": "*prepaid",
    			"Subject": "1001",
    			"SubscriberId": "1001",
    			"SysPassword": "hisPass321",
    			"SysUserName": "danb",
    			"Uuid": "388539dfd4f5cefee8f488b78c6c244b9e19138e"
    		},
    		"Weight": 0
    	}],
    	"error": null
    }

GetActionPlan
#############

Returns a list of all ActionPlans defined on user accounts:

:Hint:

    cgr> actionplan_get

*Request*

::

   {
   	"method": "ApierV1.GetActionPlan",
   	"params": [{
   		"ID": ""
   	}],
   	"id": 3
   }

*Response*

::

   {
   	"id": 3,
   	"result": [{
   			"Id": "PACKAGE_10_SHARED_A_5",
   			"AccountIDs": null,
   			"ActionTimings": [{
   				"Uuid": "93e8cb80-7dad-4efc-8d65-1e0e61ce219d",
   				"Timing": {
   					"Timing": {
   						"Years": null,
   						"Months": null,
   						"MonthDays": null,
   						"WeekDays": null,
   						"StartTime": "*asap",
   						"EndTime": ""
   					},
   					"Rating": null,
   					"Weight": 0
   				},
   				"ActionsID": "TOPUP_RST_5",
   				"Weight": 10
   			}, {
   				"Uuid": "a4ac319b-144a-49e6-b87f-8878c8adc495",
   				"Timing": {
   					"Timing": {
   						"Years": null,
   						"Months": null,
   						"MonthDays": null,
   						"WeekDays": null,
   						"StartTime": "*asap",
   						"EndTime": ""
   					},
   					"Rating": null,
   					"Weight": 0
   				},
   				"ActionsID": "TOPUP_RST_SHARED_5",
   				"Weight": 10
   			}]
   		},

   		{
   			"Id": "PACKAGE_1001",
   			"AccountIDs": {
   				"cgrates.org:1001": true
   			},
   			"ActionTimings": [{
   				"Uuid": "8261378b-aa47-45c8-a0ad-6fb4a61358a6",
   				"Timing": {
   					"Timing": {
   						"Years": null,
   						"Months": null,
   						"MonthDays": null,
   						"WeekDays": null,
   						"StartTime": "*asap",
   						"EndTime": ""
   					},
   					"Rating": null,
   					"Weight": 0
   				},
   				"ActionsID": "TOPUP_RST_5",
   				"Weight": 10
   			}, {
   				"Uuid": "a1360fae-d9e9-4a6f-9b29-c4dcdd56b266",
   				"Timing": {
   					"Timing": {
   						"Years": null,
   						"Months": null,
   						"MonthDays": null,
   						"WeekDays": null,
   						"StartTime": "*asap",
   						"EndTime": ""
   					},
   					"Rating": null,
   					"Weight": 0
   				},
   				"ActionsID": "TOPUP_RST_SHARED_5",
   				"Weight": 10
   			}, {
   				"Uuid": "f3ed64ba-a158-4302-ad46-98646cad8a8f",
   				"Timing": {
   					"Timing": {
   						"Years": null,
   						"Months": null,
   						"MonthDays": null,
   						"WeekDays": null,
   						"StartTime": "*asap",
   						"EndTime": ""
   					},
   					"Rating": null,
   					"Weight": 0
   				},
   				"ActionsID": "TOPUP_120_DST1003",
   				"Weight": 10
   			}, {
   				"Uuid": "1a5c69fb-c5f8-4852-8c66-5afd296fa5e4",
   				"Timing": {
   					"Timing": {
   						"Years": null,
   						"Months": null,
   						"MonthDays": null,
   						"WeekDays": null,
   						"StartTime": "*asap",
   						"EndTime": ""
   					},
   					"Rating": null,
   					"Weight": 0
   				},
   				"ActionsID": "TOPUP_RST_DATA_100",
   				"Weight": 10
   			}]
   		},

   		{
   			"Id": "PACKAGE_10",
   			"AccountIDs": {
   				"cgrates.org:1002": true,
   				"cgrates.org:1003": true,
   				"cgrates.org:1004": true
   			},
   			"ActionTimings": [{
   				"Uuid": "6e335f92-ae2e-4253-8809-f124a46eac06",
   				"Timing": {
   					"Timing": {
   						"Years": null,
   						"Months": null,
   						"MonthDays": null,
   						"WeekDays": null,
   						"StartTime": "*asap",
   						"EndTime": ""
   					},
   					"Rating": null,
   					"Weight": 0
   				},
   				"ActionsID": "TOPUP_RST_10",
   				"Weight": 10
   			}]
   		}, {
   			"Id": "USE_SHARED_A",
   			"AccountIDs": {
   				"cgrates.org:1007": true
   			},
   			"ActionTimings": [{
   				"Uuid": "eee41fa1-aa24-4795-b875-37213473ad3d",
   				"Timing": {
   					"Timing": {
   						"Years": null,
   						"Months": null,
   						"MonthDays": null,
   						"WeekDays": null,
   						"StartTime": "*asap",
   						"EndTime": ""
   					},
   					"Rating": null,
   					"Weight": 0
   				},
   				"ActionsID": "SHARED_A_0",
   				"Weight": 10
   			}]
   		}
   	],
   	"error": null
   }


GetActionPlan of one Package ID
###############################

Returns a list of accounts where ActionPlan for "PACKAGE_10" is allocated:

:Hint:

    cgr> actionplan_get ID="PACKAGE_10"

*Request*

::

   {
   	"method": "ApierV1.GetActionPlan",
   	"params": [{
   		"ID": "PACKAGE_10"
   	}],
   	"id": 4
   }

*Response*

::

   {
   	"id": 4,
   	"result": [{
   		"Id": "PACKAGE_10",
   		"AccountIDs": {
   			"cgrates.org:1002": true,
   			"cgrates.org:1003": true,
   			"cgrates.org:1004": true
   		},
   		"ActionTimings": [{
   			"Uuid": "6e335f92-ae2e-4253-8809-f124a46eac06",
   			"Timing": {
   				"Timing": {
   					"Years": null,
   					"Months": null,
   					"MonthDays": null,
   					"WeekDays": null,
   					"StartTime": "*asap",
   					"EndTime": ""
   				},
   				"Rating": null,
   				"Weight": 0
   			},
   			"ActionsID": "TOPUP_RST_10",
   			"Weight": 10
   		}]
   	}],
   	"error": null
   }


User Indexes
############

:Hint:

    cgr> user_indexes

*Request*

::

    {
    	"method": "UsersV1.GetIndexes",
    	"params": [""],
    	"id": 2
    }

*Response*

::

    {
    	"id": 2,
    	"result": {
    		"Uuid:27f37edec0670fa34cf79076b80ef5021e39c5b5": ["cgrates.org:1002"],
    		"Uuid:388539dfd4f5cefee8f488b78c6c244b9e19138e": ["cgrates.org:1001"]
    	},
    	"error": null
    }

CDR Stats for Queues
#######################

Return list of Queue IDs

:Hint:

    cgr> cdrstats_queueids

*Request*

::

    {
    	"method": "CDRStatsV1.GetQueueIds",
    	"params": [""],
    	"id": 8
    }

*Response*

::

    {
    	"id": 8,
    	"result": [
    		"CDRST_1003",
    		"CDRST1",
    		"CDRST_1001",
    		"CDRST_1002",
    		"STATS_SUPPL1",
    		"STATS_SUPPL2"
    	],
    	"error": null
    }


Debit Air Time
##############

:Hint:

    cgr> debit Tenant="cgrates.org" Account="1001" CallDuration=500

*Request*

::

    {
    	"method": "Responder.Debit",
    	"params": [{
    		"Direction": "*out",
    		"Category": "",
    		"Tenant": "cgrates.org",
    		"Subject": "",
    		"Account": "1001",
    		"Destination": "",
    		"TimeStart": "0001-01-01T00:00:00Z",
    		"TimeEnd": "0001-01-01T00:00:00Z",
    		"LoopIndex": 0,
    		"DurationIndex": 0,
    		"FallbackSubject": "",
    		"RatingInfos": null,
    		"Increments": null,
    		"TOR": "",
    		"ExtraFields": null,
    		"MaxRate": 0,
    		"MaxRateUnit": 0,
    		"MaxCostSoFar": 0,
    		"CgrID": "",
    		"RunID": "",
    		"ForceDuration": false,
    		"PerformRounding": false,
    		"DryRun": false,
    		"DenyNegativeAccount": false
    	}],
    	"id": 16
    }

*Response*

::

    {
    	"id": 16,
    	"result": {
    		"Direction": "*out",
    		"Category": "",
    		"Tenant": "cgrates.org",
    		"Subject": "1001",
    		"Account": "1001",
    		"Destination": "",
    		"TOR": "",
    		"Cost": 0,
    		"Timespans": null,
    		"RatedUsage": 0,
    		"AccountSummary": {
    			"Tenant": "cgrates.org",
    			"ID": "1001",
    			"BalanceSummaries": [{
    				"UUID": "a6fc6e96-de69-445b-8456-cebd78a1b43d",
    				"ID": "a6fc6e96-de69-445b-8456-cebd78a1b43d",
    				"Type": "*monetary",
    				"Value": 5,
    				"Disabled": false
    			}, {
    				"UUID": "9df5d845-e411-4edd-971c-d98dbb926054",
    				"ID": "9df5d845-e411-4edd-971c-d98dbb926054",
    				"Type": "*monetary",
    				"Value": 25,
    				"Disabled": false
    			}, {
    				"UUID": "4a4d07c8-9548-415d-a029-7e369bf02f60",
    				"ID": "4a4d07c8-9548-415d-a029-7e369bf02f60",
    				"Type": "*voice",
    				"Value": 120,
    				"Disabled": false
    			}, {
    				"UUID": "8d867c57-31b4-407d-afc7-fb4dc359ae4d",
    				"ID": "8d867c57-31b4-407d-afc7-fb4dc359ae4d",
    				"Type": "*voice",
    				"Value": 90,
    				"Disabled": false
    			}, {
    				"UUID": "66009d4e-25ed-47d6-8dfa-ef3c501fd1b0",
    				"ID": "66009d4e-25ed-47d6-8dfa-ef3c501fd1b0",
    				"Type": "*data",
    				"Value": 102400,
    				"Disabled": false
    			}],
    			"AllowNegative": false,
    			"Disabled": false
    		}
    	},
    	"error": null
    }

Set Balance for Outbound Calls
##############################

:Hint:

    cgr> balance_set Tenant="cgrates.org" Account="1001" BalanceType="\*voice" Directions="\*out" Value=100 BalanceID="8d867c57-31b4-407d-afc7-fb4dc359ae4d"

*Request*

::

    {
    	"method": "ApierV1.SetBalance",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"Account": "1001",
    		"BalanceType": "*voice",
    		"BalanceUUID": null,
    		"BalanceID": "8d867c57-31b4-407d-afc7-fb4dc359ae4d",
    		"Directions": "*out",
    		"Value": 100,
    		"ExpiryTime": null,
    		"RatingSubject": null,
    		"Categories": null,
    		"DestinationIds": null,
    		"TimingIds": null,
    		"Weight": null,
    		"SharedGroups": null,
    		"Blocker": null,
    		"Disabled": null
    	}],
    	"id": 18
    }

*Response*

::

    {
    	"id": 18,
    	"result": "OK",
    	"error": null
    }

Set Balance for Inbound Calls
#############################

:Hint:

    cgr> balance_set Tenant="cgrates.org" Account="1001" BalanceType="\*voice" Directions="\*in" Value=600 BalanceID="9d867c57-31b4-407d-afc7-fb4dc359ae4d"

*Request*

::

    {
    	"method": "ApierV1.SetBalance",
    	"params": [{
    		"Tenant": "cgrates.org",
    		"Account": "1001",
    		"BalanceType": "*voice",
    		"BalanceUUID": null,
    		"BalanceID": "9d867c57-31b4-407d-afc7-fb4dc359ae4d",
    		"Directions": "*in",
    		"Value": 600,
    		"ExpiryTime": null,
    		"RatingSubject": null,
    		"Categories": null,
    		"DestinationIds": null,
    		"TimingIds": null,
    		"Weight": null,
    		"SharedGroups": null,
    		"Blocker": null,
    		"Disabled": null
    	}],
    	"id": 28
    }

*Response*

::

    {
    	"id": 28,
    	"result": "OK",
    	"error": null
    }
