ApierV1.SetTPRate
+++++++++++++++

Creates a new rate within a tariff plan.

**Request**:

 Data:
  ::

   type TPRate struct {
	TPid      string     // Tariff plan id
	RateId    string     // Rate id
	RateSlots []RateSlot // One or more RateSlots
   }

   type RateSlot struct {
	ConnectFee         float64 // ConnectFee applied once the call is answered
	Rate               float64 // Rate applied
	RateUnit           string  //  Number of billing units this rate applies to
	RateIncrement      string  // This rate will apply in increments of duration
	GroupIntervalStart string  // Group position
	RoundingMethod     string  // Use this method to round the cost
	RoundingDecimals   int     // Round the cost number of decimals
   }

 Mandatory parameters: ``[]string{"TPid", "RateId", "ConnectFee", "RateSlots"}``

 *JSON sample*:
  ::

   {
    "id": 2,
    "method": "ApierV1.SetTPRate",
    "params": [
        {
            "RateId": "1CENTPERSEC",
            "RateSlots": [
                {
                    "ConnectFee": 0,
                    "GroupIntervalStart": "0",
                    "Rate": 0.01,
                    "RateIncrement": "1s",
                    "RateUnit": "1s",
                    "RoundingDecimals": 4,
                    "RoundingMethod": "*middle",
                    "Weight": 10
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
    "id": 0, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/RateId already exists in StorDb.


ApierV1.GetTPRate
+++++++++++++++

Queries specific rate on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPRate struct {
	TPid   string // Tariff plan id
	RateId string // Rate id
   }

 Mandatory parameters: ``[]string{"TPid", "RateId"}``

 *JSON sample*:
  ::

   {
    "id": 3,
    "method": "ApierV1.GetTPRate",
    "params": [
        {
            "RateId": "1CENTPERSEC",
            "TPid": "CGR_API_TESTS"
        }
    ]
   }
   
**Reply**:

 Data:
  ::

   type TPRate struct {
	TPid      string     // Tariff plan id
	RateId    string     // Rate id
	RateSlots []RateSlot // One or more RateSlots
   }

   type RateSlot struct {
	ConnectFee         float64 // ConnectFee applied once the call is answered
	Rate               float64 // Rate applied
	RateUnit           string  //  Number of billing units this rate applies to
	RateIncrement      string  // This rate will apply in increments of duration
	GroupIntervalStart string  // Group position
	RoundingMethod     string  // Use this method to round the cost
	RoundingDecimals   int     // Round the cost number of decimals
   }

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 3,
    "result": {
        "RateId": "1CENTPERSEC",
        "RateSlots": [
            {
                "ConnectFee": 0,
                "GroupIntervalStart": "0",
                "Rate": 0.01,
                "RateIncrement": "1s",
                "RateUnit": "1s",
                "RoundingDecimals": 4,
                "RoundingMethod": "*middle"
            }
        ],
        "TPid": "CGR_API_TESTS"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested rate id not found.


ApierV1.GetTPRateIds
++++++++++++++++++

Queries rate identities on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPRateIds struct {
	TPid string // Tariff plan id
   }

 Mandatory parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 4,
    "method": "ApierV1.GetTPRateIds",
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
    "id": 4,
    "result": [
        "1CENTPERSEC"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested tariff plan not found.


