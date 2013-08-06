ApierV1.SetTPDestination
++++++++++++++++++++++

Creates a new destination within a tariff plan id.

**Request**:

 Data:
  ::

   type ApierTPDestination struct {
	TPid          string   // Tariff plan id
	DestinationId string   // Destination id
	Prefixes      []string // Prefixes attached to this destination
   }

 Required parameters: ``[]string{"TPid", "DestinationId", "Prefixes"}``

 *JSON sample*:
  ::

   {
     "id": 2, 
     "method": "ApierV1.SetTPDestination", 
     "params": [
        {
            "DestinationId": "FIST_DST2", 
            "Prefixes": [
                "123", 
                "345"
            ], 
            "TPid": "FIST_TP"
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

 ``DUPLICATE`` - The specified combination of TPid/DestinationId already exists in StorDb.


ApierV1.GetTPDestination
++++++++++++++++++++++

Queries a specific destination.

**Request**:

 Data:
  ::

   type AttrGetTPDestination struct {
	TPid          string // Tariff plan id
	DestinationId string // Destination id
   }

 Required parameters: ``[]string{"TPid", "DestinationId"}``

 *JSON sample*:
  ::

   {
    "id": 0, 
    "method": "ApierV1.GetTPDestination", 
    "params": [
        {
            "DestinationId": "FIRST_DST2", 
            "TPid": "FIRST_TP"
        }
    ]
   }

**Reply**:

 Data:
  ::

   type ApierTPDestination struct {
	TPid          string   // Tariff plan id
	DestinationId string   // Destination id
	Prefixes      []string // Prefixes attached to this destination
   }

 *JSON sample*:
  ::

   {
    "error": null, 
    "id": 0, 
    "result": {
        "TPid":"FIST_TP",
        "DestinationId": "FIST_DST2", 
        "Prefixes": [
            "123", 
            "345"
        ]
    }
   }


**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested destination id not found.


ApierV1.GetTPDestinationIds
+++++++++++++++++++++++++

Queries destination identities on specific tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPDestinationIds struct {
	TPid string // Tariff plan id
   }

 Required parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 1, 
    "method": "ApierV1.GetTPDestinationIds", 
    "params": [
        {
            "TPid": "FIST_TP"
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
    "id": 1, 
    "result": [
        "FIST_DST", 
        "FIST_DST1", 
        "FIST_DST2", 
        "FIST_DST3", 
        "FIST_DST4"
    ]
   }



**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested tariff plan not found.

