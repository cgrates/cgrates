Apier.SetTPActionTriggers
+++++++++++++++++++++++++

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
	BalanceId      string  // Id of the balance this trigger monitors
	Direction      string  // Traffic direction
	ThresholdType  string  // This threshold type
	ThresholdValue float64 // Threshold
	DestinationId  string  // Id of the destination profile
	ActionsId      string  // Actions which will execute on threshold reached
	Weight         float64 // weight
   }

 Mandatory parameters: ``[]string{"TPid", "ActionTriggersId","BalanceId", "Direction", "ThresholdType", "ThresholdValue", "ActionsId", "Weight"}``

 *JSON sample*:
  ::

   {
    "id": 2, 
    "method": "Apier.SetTPActionTriggers", 
    "params": [
        {
            "ActionTriggers": [
                {
                    "ActionsId": "ACTION_1", 
                    "BalanceId": "MONETARY", 
                    "DestinationId": "", 
                    "Direction": "OUT", 
                    "ThresholdType": "MIN_BALANCE", 
                    "ThresholdValue": 5, 
                    "Weight": 10
                }
            ], 
            "ActionTriggersId": "SAMPLE_ATS_1", 
            "TPid": "SAMPLE_TP_2"
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
    "id": 2, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/ActionTriggersId already present in StorDb.


Apier.GetTPActionTriggers
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
    "id": 0, 
    "method": "Apier.GetTPActionTriggers", 
    "params": [
        {
            "ActionTriggersId": "SAMPLE_ATS_1", 
            "TPid": "SAMPLE_TP_2"
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
	BalanceId      string  // Id of the balance this trigger monitors
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
    "id": 0, 
    "result": {
        "ActionTriggers": [
            {
                "ActionsId": "ACTION_1", 
                "BalanceId": "MONETARY", 
                "DestinationId": "", 
                "Direction": "OUT", 
                "ThresholdType": "MIN_BALANCE", 
                "ThresholdValue": 5, 
                "Weight": 10
            }
        ], 
        "ActionTriggersId": "SAMPLE_ATS_1", 
        "TPid": "SAMPLE_TP_2"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested ActionTriggersId profile not found.


Apier.GetTPActionTriggerIds
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
    "id": 5, 
    "method": "Apier.GetTPActionTriggerIds", 
    "params": [
        {
            "TPid": "SAMPLE_TP_2"
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
    "id": 5, 
    "result": [
        "SAMPLE_ATS_1",
        "SAMPLE_ATS_2"
    ]
}

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - There are no ActionTriggers profiles defined on the selected TPid.
