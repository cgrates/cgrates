ApierV1.SetTPDestinationRate
++++++++++++++++++++++++++


Creates a new DestinationRate profile within a tariff plan.

**Request**:

 Data:
  ::

   type TPDestinationRate struct {
	TPid              string // Tariff plan id
	DestinationRateId string // DestinationRate profile id
	DestinationRates     []DestinationRate // Set of destinationid-rateid bindings
   }

   type DestinationRate struct {
	DestinationId string // The destination identity
	RateId		string // The rate identity
   }

 Mandatory parameters: ``[]string{"TPid", "DestinationRateId", "DestinationRates"}``

 *JSON sample*:
  ::

   {
    "id": 7,
    "method": "ApierV1.SetTPDestinationRate",
    "params": [
        {
            "DestinationRateId": "DR_1CENTPERSEC",
            "DestinationRates": [
                {
                    "DestinationId": "FS_USERS",
                    "RateId": "1CENTPERSEC"
                }
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
    "id": 7, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/DestinationRateId already exists in StorDb.


ApierV1.GetTPDestinationRate
+++++++++++++++

Queries specific DestinationRate profile on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPDestinationRate struct {
	TPid   string // Tariff plan id
	DestinationRateId string // Rate id
   }

 Mandatory parameters: ``[]string{"TPid", "DestinationRateId"}``

 *JSON sample*:
  ::

   {
    "id": 8,
    "method": "ApierV1.GetTPDestinationRate",
    "params": [
        {
            "DestinationRateId": "DR_1CENTPERSEC",
            "TPid": "CGR_API_TESTS"
        }
    ]
   }
   
**Reply**:

 Data:
  ::

   type TPDestinationRate struct {
	TPid              string // Tariff plan id
	DestinationRateId string // DestinationRate profile id
	DestinationRates     []DestinationRate // Set of destinationid-rateid bindings
   }

   type DestinationRate struct {
	DestinationId string // The destination identity
	RateId		string // The rate identity
   }

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 8,
    "result": {
        "DestinationRateId": "DR_1CENTPERSEC",
        "DestinationRates": [
            {
                "DestinationId": "FS_USERS",
                "RateId": "1CENTPERSEC"
            }
        ],
        "TPid": "CGR_API_TESTS"
    }
  }


**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested DestinationRate id not found.


ApierV1.GetTPDestinationRateIds
+++++++++++++++++++++++++++++

Queries DestinationRate identities on specific tariff plan.

**Request**:

 Data:
  ::

   type AttrTPDestinationRateIds struct {
	TPid string // Tariff plan id
   }

 Mandatory parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 9,
    "method": "ApierV1.GetTPDestinationRateIds",
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
    "id": 9,
    "result": [
        "DR_1CENTPERSEC"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested tariff plan not found.

