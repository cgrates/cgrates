ApierV1.SetTPActionTriggers
+++++++++++++++++++++++++++

Creates a new ActionTriggers profile within a tariff plan.

**Request**:

 Data:
  ::

   type ApiTPActionTriggers struct {
	TPid             string             // Tariff plan id
	ActionTriggersId string             // Profile id
	ActionTriggers   []ApiActionTrigger // Set of triggers grouped in this profile

   }

   type ApiActionTrigger struct {
	BalanceType      string  // Id of the balance this trigger monitors
	Direction      string  // Traffic direction
	ThresholdType  string  // This threshold type
	ThresholdValue float64 // Threshold
	DestinationId  string  // Id of the destination profile
	ActionsId      string  // Actions which will execute on threshold reached
	Weight         float64 // weight
   }

 Mandatory parameters: ``[]string{"TPid", "ActionTriggersId","BalanceType", "Direction", "ThresholdType", "ThresholdValue", "ActionsId", "Weight"}``

 *JSON sample*:
  ::

   {
    "id": 45,
    "method": "ApierV1.SetTPActionTriggers",
    "params": [
        {
            "ActionTriggers": [
                {
                    "ActionsId": "LOG_BALANCE",
                    "BalanceType": "*monetary",
                    "DestinationId": "",
                    "Direction": "*out",
                    "ThresholdType": "*min_balance",
                    "ThresholdValue": 2,
                    "Weight": 10
                },
                {
                    "ActionsId": "LOG_BALANCE",
                    "BalanceType": "*monetary",
                    "DestinationId": "",
                    "Direction": "*out",
                    "ThresholdType": "*max_balance",
                    "ThresholdValue": 20,
                    "Weight": 10
                },
                {
                    "ActionsId": "LOG_BALANCE",
                    "BalanceType": "*monetary",
                    "DestinationId": "FS_USERS",
                    "Direction": "*out",
                    "ThresholdType": "*max_counter",
                    "ThresholdValue": 15,
                    "Weight": 10
                }
            ],
            "ActionTriggersId": "STANDARD_TRIGGERS",
            "TPid": "CGR_API_TESTS"
        }
    ]
   }

**Reply**:

 Data:
  ::

   string

 Possible answers:
  ``OK`` - Success.

 *JSON sample*:
  ::

   {
    "error": null, 
    "id": 45, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/ActionTriggersId already present in StorDb.


ApierV1.GetTPActionTriggers
+++++++++++++++++++++++++

Queries specific ActionTriggers profile on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPActionTriggers struct {
	TPid             string // Tariff plan id
	ActionTriggersId string // ActionTrigger id
   }

 Mandatory parameters: ``[]string{"TPid", "ActionTriggersId"}``

 *JSON sample*:
  ::

   {
    "id": 46,
    "method": "ApierV1.GetTPActionTriggers",
    "params": [
        {
            "ActionTriggersId": "STANDARD_TRIGGERS",
            "TPid": "CGR_API_TESTS"
        }
    ]
   }
 
**Reply**:

 Data:
  ::

   type ApiTPActionTriggers struct {
	TPid             string             // Tariff plan id
	ActionTriggersId string             // Profile id
	ActionTriggers   []ApiActionTrigger // Set of triggers grouped in this profile

   }

   type ApiActionTrigger struct {
	BalanceType      string  // Id of the balance this trigger monitors
	Direction      string  // Traffic direction
	ThresholdType  string  // This threshold type
	ThresholdValue float64 // Threshold
	DestinationId  string  // Id of the destination profile
	ActionsId      string  // Actions which will execute on threshold reached
	Weight         float64 // weight
   }

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 46,
    "result": {
        "ActionTriggers": [
            {
                "ActionsId": "LOG_BALANCE",
                "BalanceType": "*monetary",
                "DestinationId": "",
                "Direction": "*out",
                "ThresholdType": "*min_balance",
                "ThresholdValue": 2,
                "Weight": 10
            },
            {
                "ActionsId": "LOG_BALANCE",
                "BalanceType": "*monetary",
                "DestinationId": "",
                "Direction": "*out",
                "ThresholdType": "*max_balance",
                "ThresholdValue": 20,
                "Weight": 10
            },
            {
                "ActionsId": "LOG_BALANCE",
                "BalanceType": "*monetary",
                "DestinationId": "FS_USERS",
                "Direction": "*out",
                "ThresholdType": "*max_counter",
                "ThresholdValue": 15,
                "Weight": 10
            }
        ],
        "ActionTriggersId": "STANDARD_TRIGGERS",
        "TPid": "CGR_API_TESTS"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested ActionTriggersId profile not found.


ApierV1.GetTPActionTriggerIds
+++++++++++++++++++++++++++

Queries ActionTriggers identities on specific tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPActionTriggerIds struct {
	TPid string // Tariff plan id
   }

 Mandatory parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 47,
    "method": "ApierV1.GetTPActionTriggerIds",
    "params": [
        {
            "TPid": "CGR_API_TESTS"
        }
    ]
   }

**Reply**:

 Data:
  ::

   []string

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 47,
    "result": [
        "STANDARD_TRIGGERS"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - There are no ActionTriggers profiles defined on the selected TPid.
