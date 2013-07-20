Apier.SetTPActionTimings
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
    "id": 7, 
    "method": "Apier.SetTPActionTimings", 
    "params": [
        {
            "ActionTimings": [
                {
                    "ActionsId": "SAMPLE_ACTIONS", 
                    "TimingId": "ALL_TIME", 
                    "Weight": 10
                }, 
                {
                    "ActionsId": "SAMPLE_ACTIONS2", 
                    "TimingId": "ALL_TIME", 
                    "Weight": 10
                }
            ], 
            "ActionTimingsId": "SAMPLE_AT3", 
            "TPid": "SAMPLE_TP_1"
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
    "id": 7, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/ActionTimingsId already present in StorDb.


Apier.GetTPActionTimings
++++++++++++++++++++++++

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
    "id": 8, 
    "method": "Apier.GetTPActionTimings", 
    "params": [
        {
            "ActionTimingsId": "SAMPLE_AT3", 
            "TPid": "SAMPLE_TP_1"
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
    "id": 8, 
    "result": {
        "ActionTimings": [
            {
                "ActionsId": "SAMPLE_ACTIONS", 
                "TimingId": "ALL_TIME", 
                "Weight": 10
            }, 
            {
                "ActionsId": "SAMPLE_ACTIONS2", 
                "TimingId": "ALL_TIME", 
                "Weight": 10
            }
        ], 
        "ActionTimingsId": "SAMPLE_AT3", 
        "TPid": "SAMPLE_TP_1"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested ActionTimings profile not found.


Apier.GetTPActionTimingIds
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
    "id": 9, 
    "method": "Apier.GetTPActionTimingIds", 
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
    "id": 9, 
    "result": [
        "SAMPLE_AT", 
        "SAMPLE_AT2", 
        "SAMPLE_AT3"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - There are no ActionTimings profiles defined on the selected TPid.
