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
    "id": 6,
    "method": "ApierV1.SetTPDestination",
    "params": [
        {
            "DestinationId": "FS_USERS",
            "Prefixes": [
                "10"
            ],
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
    "id": 6,
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
    "id": 7,
    "method": "ApierV1.GetTPDestination",
    "params": [
        {
            "DestinationId": "FS_USERS",
            "TPid": "CGR_API_TESTS"
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
    "id": 7,
    "result": {
        "DestinationId": "FS_USERS",
        "Prefixes": [
            "10"
        ],
        "TPid": "CGR_API_TESTS"
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
    "id": 8,
    "method": "ApierV1.GetTPDestinationIds",
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
    "id": 8,
    "result": [
        "FS_USERS"
    ]
  }



**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested tariff plan not found.

