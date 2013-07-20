Apier.SetTPAccountActions
+++++++++++++++++++++++++

Creates a new AccountActions profile within a tariff plan.

**Request**:

 Data:
  ::

   type ApiTPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // AccountActions id
	Tenant           string // Tenant's Id
	Account          string // Account name
	Direction        string // Traffic direction
	ActionTimingsId  string // Id of ActionTimings profile to use
	ActionTriggersId string // Id of ActionTriggers profile to use
   }

 Mandatory parameters: ``[]string{"TPid", "AccountActionsId","Tenant","Account","Direction","ActionTimingsId","ActionTriggersId"}``

 *JSON sample*:
  ::

   {
    "id": 2, 
    "method": "Apier.SetTPAccountActions", 
    "params": [
        {
            "Account": "ACNT1", 
            "AccountActionsId": "AA_SAMPLE_2", 
            "ActionTimingsId": "SAMPLE_AT_1", 
            "ActionTriggersId": "SAMPLE_ATRS_1", 
            "Direction": "OUT", 
            "TPid": "SAMPLE_TP_1", 
            "Tenant": "TENANT1"
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

 ``DUPLICATE`` - The specified combination of TPid/AccountActionsId already present in StorDb.


Apier.GetTPAccountActions
+++++++++++++++++++++++++

Queries specific AccountActions profile on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // AccountActions id
   }

 Mandatory parameters: ``[]string{"TPid", "AccountActionsId"}``

 *JSON sample*:
  ::

   {
    "id": 3, 
    "method": "Apier.GetTPAccountActions", 
    "params": [
        {
            "AccountActionsId": "AA_SAMPLE_2", 
            "TPid": "SAMPLE_TP_1"
        }
    ]
   }
 
**Reply**:

 Data:
  ::

   type ApiTPAccountActions struct {
	TPid             string // Tariff plan id
	AccountActionsId string // AccountActions id
	Tenant           string // Tenant's Id
	Account          string // Account name
	Direction        string // Traffic direction
	ActionTimingsId  string // Id of ActionTimings profile to use
	ActionTriggersId string // Id of ActionTriggers profile to use
   }

 *JSON sample*:
  ::

   {
    "error": null, 
    "id": 3, 
    "result": {
        "Account": "ACNT1", 
        "AccountActionsId": "AA_SAMPLE_2", 
        "ActionTimingsId": "SAMPLE_AT_1", 
        "ActionTriggersId": "SAMPLE_ATRS_1", 
        "Direction": "OUT", 
        "TPid": "SAMPLE_TP_1", 
        "Tenant": "TENANT1"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested AccountActions profile not found.


Apier.GetTPAccountActionIds
+++++++++++++++++++++++++++

Queries AccountActions identities on specific tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPAccountActionIds struct {
	TPid string // Tariff plan id
   }

 Mandatory parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 4, 
    "method": "Apier.GetTPAccountActionIds", 
    "params": [
        {
            "TPid": "SAMPLE_TP_1"
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
    "id": 4, 
    "result": [
        "AA_SAMPLE_1", 
        "AA_SAMPLE_2"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - There are no AccountAction profiles defined on the selected TPid.


