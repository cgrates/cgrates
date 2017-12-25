4. CDR Management
=================

4.1 Export CDRs
---------------

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

4.2 List all CDRs
-----------------

:Hint:
    db.getCollection('cdrs').find({})

4.2.1 Filter based on 'cgrid'
#############################

:Hint:
    db.cdrs.find({"cgrid":"84bde1fd133f70572e05e699ea2f1de201e18269", "runid":"\*default"})

4.2.2 Filter calls from 1001 to 1002
####################################

:Hint:
    db.cdrs.find({"account":"1001", "destination":"1002"})

4.2.3 Filter calls from 1003 to 1002
####################################

:Hint:
    db.cdrs.find({"account":"1003", "destination":"1002"})

4.2.4 Filter calls on setup time
################################

:Hint:

    db.cdrs.find({"setuptime" : ISODate("2017-12-11T23:38:57.000Z")})


4.3 CDR Stats for Queues
------------------------

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
