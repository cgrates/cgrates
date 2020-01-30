4. CDR Management
=================

4.1 Export CDRs
---------------

:Hint:
    cgr > cdrs_export CdrFormat="csv" ExportDir="/tmp"

*Request*

::

    {
    	"method": "APIerSv1.ExportCDRs",
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


"/var/spool/cgrates/cdre/cdre_1513199075.csv" is the destination cdr file in csv format.


4.2 CDR Stats for Queues
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
