2. User Life Cycle
==================

Following steps will cover use-case "User Life cycle"

2.1 User Management
-------------------

2.1.1 Create User Account
#########################

:Hint:
    cgr> account_set Tenant="cgrates.org" Account="1003" ActionPlanIDs=["PACKAGE_10"] ActionTriggerIDs=["STANDARD_TRIGGERS"]

*Request*

::

    {
    	"method": "APIerSv2.SetAccount",
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

2.1.2 Get User Account
######################

:Hint:
    cgr> accounts Tenant="cgrates.org" AccountIds=["1003"]

*Request*

::

    {
    	"method": "APIerSv2.GetAccounts",
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

2.1.3 Remove User Account
#########################

:Hint:
    cgr> account_remove Tenant="cgrates.org" Account="1003"

*Request*

::

    {
    	"method": "APIerSv1.RemoveAccount",
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

2.1.4 Get Users Profile
#######################

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

2.1.5 Get Profile UserName 1001
###############################

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

2.1.6 Get Action Plans
######################

Returns a list of all ActionPlans defined on user accounts:

:Hint:

    cgr> actionplan_get

*Request*

::

   {
   	"method": "APIerSv1.GetActionPlan",
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


2.1.7 Get Action Plans of one Package ID
########################################

Returns a list of accounts where ActionPlan for "PACKAGE_10" is allocated:

:Hint:

    cgr> actionplan_get ID="PACKAGE_10"

*Request*

::

   {
   	"method": "APIerSv1.GetActionPlan",
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


2.1.8 User Indexes
##################

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
