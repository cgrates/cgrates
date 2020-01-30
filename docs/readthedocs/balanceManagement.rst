3. Balance Management
=====================

3.1 Balance Set
---------------

 replaces existing value of BalanceID '23456' with value 12 for account 1003 belongs to tenant 'cgrates.org'

:Hint:
    cgr> balance_set Tenant="cgrates.org" Account="1003" Direction="\*out" Value=12 BalanceID="23456"

*Request*

::

    {
    	"method": "APIerSv1.SetBalance",
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

3.2 Balance Add
---------------

 adds 10 cent to account=1003 where tenant=cgrates.org

:Hint:
    cgr> balance_add Tenant="cgrates.org" Account="1003" BalanceId="123456" Value=10

*Request*

::

    {
    	"method": "APIerSv1.AddBalance",
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


3.3 Balance Debit
-----------------

 deducts 5 cents from account 1003 of tenant cgrates.org

:Hint:
    cgr> balance_debit Tenant="cgrates.org" Account="1003" BalanceId="23456" Value=5 BalanceType="\*monetary"

*Request*

::

    {
    	"method": "APIerSv1.DebitBalance",
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


3.4 Get Remaining Balance
-------------------------

Sum of BalanceMap.Value resulted from APIerSv2.GetAccounts request


3.5 Debit Air Time (TBV)
------------------------

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
    		"ToR": "",
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
    		"ToR": "",
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

3.6 Set Balance for Outbound Calls
----------------------------------

:Hint:

    cgr> balance_set Tenant="cgrates.org" Account="1001" BalanceType="\*voice" Directions="\*out" Value=100 BalanceID="8d867c57-31b4-407d-afc7-fb4dc359ae4d"

*Request*

::

    {
    	"method": "APIerSv1.SetBalance",
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

3.7 Set Balance for Inbound Calls
---------------------------------

:Hint:

    cgr> balance_set Tenant="cgrates.org" Account="1001" BalanceType="\*voice" Directions="\*in" Value=600 BalanceID="9d867c57-31b4-407d-afc7-fb4dc359ae4d"

*Request*

::

    {
    	"method": "APIerSv1.SetBalance",
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
