ApierV1.SetTPAccountActions
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
    "id": 48,
    "method": "ApierV1.SetTPAccountActions",
    "params": [
        {
            "Account": "1005",
            "AccountActionsId": "AA_1005",
            "ActionTimingsId": "AT_FS10",
            "ActionTriggersId": "STANDARD_TRIGGERS",
            "Direction": "*out",
            "TPid": "CGR_API_TESTS",
            "Tenant": "cgrates.org"
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
    "id": 48, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/AccountActionsId already present in StorDb.


ApierV1.GetTPAccountActions
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
    "id": 49,
    "method": "ApierV1.GetTPAccountActions",
    "params": [
        {
            "AccountActionsId": "AA_1005",
            "TPid": "CGR_API_TESTS"
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
    "id": 49,
    "result": {
        "Account": "1005",
        "AccountActionsId": "AA_1005",
        "ActionTimingsId": "AT_FS10",
        "ActionTriggersId": "STANDARD_TRIGGERS",
        "Direction": "*out",
        "TPid": "CGR_API_TESTS",
        "Tenant": "cgrates.org"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested AccountActions profile not found.


ApierV1.GetTPAccountActionIds
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
    "id": 50,
    "method": "ApierV1.GetTPAccountActionIds",
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
    "id": 50,
    "result": [
        "AA_1005"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - There are no AccountAction profiles defined on the selected TPid.


