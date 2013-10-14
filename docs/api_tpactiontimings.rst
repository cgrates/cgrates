ApierV1.SetTPActionTimings
++++++++++++++++++++++++

Creates a new ActionTimings profile within a tariff plan.

**Request**:

 Data:
  ::

   type ApiTPActionTimings struct {
	TPid            string         // Tariff plan id
	ActionTimingsId string         // ActionTimings id
	ActionTimings   []ApiActionTiming // Set of ActionTiming bindings this profile will group
   }

   type ApiActionTiming struct {
	ActionsId string  // Actions id
	TimingId  string  // Timing profile id
	Weight    float64 // Binding's weight
   }

 Mandatory parameters: ``[]string{"TPid", "ActionTimingsId", "ActionTimings", "ActionsId", "TimingId", "Weight"}``

 *JSON sample*:
  ::

   {
    "id": 42,
    "method": "ApierV1.SetTPActionTimings",
    "params": [
        {
            "ActionTimings": [
                {
                    "ActionsId": "TOPUP_10",
                    "TimingId": "ASAP",
                    "Weight": 10
                }
            ],
            "ActionTimingsId": "AT_FS10",
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
    "id": 42, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/ActionTimingsId already present in StorDb.


ApierV1.GetTPActionTimings
++++++++++++++++++++++++++

Queries specific ActionTimings profile on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPActionTimings struct {
	TPid      string // Tariff plan id
	ActionTimingsId string // ActionTimings id
   }

 Mandatory parameters: ``[]string{"TPid", "ActionTimingsId"}``

 *JSON sample*:
  ::

   {
    "id": 43,
    "method": "ApierV1.GetTPActionTimings",
    "params": [
        {
            "ActionTimingsId": "AT_FS10",
            "TPid": "CGR_API_TESTS"
        }
    ]
   }
 
**Reply**:

 Data:
  ::

   type ApiTPActionTimings struct {
	TPid            string         // Tariff plan id
	ActionTimingsId string         // ActionTimings id
	ActionTimings   []ApiActionTiming // Set of ActionTiming bindings this profile will group
   }

   type ApiActionTiming struct {
	ActionsId string  // Actions id
	TimingId  string  // Timing profile id
	Weight    float64 // Binding's weight
   }

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 43,
    "result": {
        "ActionTimings": [
            {
                "ActionsId": "TOPUP_10",
                "TimingId": "ASAP",
                "Weight": 10
            }
        ],
        "ActionTimingsId": "AT_FS10",
        "TPid": "CGR_API_TESTS"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested ActionTimings profile not found.


ApierV1.GetTPActionTimingIds
++++++++++++++++++++++++++

Queries ActionTimings identities on specific tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPActionTimingIds struct {
	TPid string // Tariff plan id
   }

 Mandatory parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 44,
    "method": "ApierV1.GetTPActionTimingIds",
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
    "id": 44,
    "result": [
        "AT_FS10"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - There are no ActionTimings profiles defined on the selected TPid.
