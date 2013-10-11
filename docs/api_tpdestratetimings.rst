ApierV1.SetTPDestRateTiming
+++++++++++++++++++++++++

Creates a new DestinationRateTiming profile within a tariff plan.

**Request**:

 Data:
  ::

   type TPDestRateTiming struct {
	TPid             string           // Tariff plan id
	DestRateTimingId string           // DestinationRate profile id
	DestRateTimings  []DestRateTiming // Set of destinationid-rateid bindings
   }

   type DestRateTiming struct {
	DestRatesId string  // The DestinationRate identity
	TimingId    string  // The timing identity
	Weight      float64 // Binding priority taken into consideration when more DestinationRates are active on a time slot
   }

 Mandatory parameters: ``[]string{"TPid", "DestRateTimingId", "DestRateTimings"}``

 *JSON sample*:
  ::

   {
    "id": 10,
    "method": "ApierV1.SetTPDestRateTiming",
    "params": [
        {
            "DestRateTimingId": "DRT_1CENTPERSEC",
            "DestRateTimings": [
                {
                    "DestRatesId": "DR_1CENTPERSEC",
                    "TimingId": "ALWAYS",
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
    "id": 10, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/DestRateTimingId already exists in StorDb.


ApierV1.GetTPDestRateTiming
+++++++++++++++++++++++++

Queries specific DestRateTiming profile on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPDestRateTiming struct {
	TPid             string // Tariff plan id
	DestRateTimingId string // Rate id
   }

 Mandatory parameters: ``[]string{"TPid", "DestRateTimingId"}``

 *JSON sample*:
  ::

   {
    "id": 11,
    "method": "ApierV1.GetTPDestRateTiming",
    "params": [
        {
            "DestRateTimingId": "DRT_1CENTPERSEC",
            "TPid": "CGR_API_TESTS"
        }
    ]
   }
   
**Reply**:

 Data:
  ::

   type TPDestRateTiming struct {
	TPid             string           // Tariff plan id
	DestRateTimingId string           // DestinationRate profile id
	DestRateTimings  []DestRateTiming // Set of destinationid-rateid bindings
   }

   type DestRateTiming struct {
	DestRatesId string  // The DestinationRate identity
	TimingId    string  // The timing identity
	Weight      float64 // Binding priority taken into consideration when more DestinationRates are active on a time slot
   }

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 11,
    "result": {
        "DestRateTimingId": "DRT_1CENTPERSEC",
        "DestRateTimings": [
            {
                "DestRatesId": "DR_1CENTPERSEC",
                "TimingId": "ALWAYS",
                "Weight": 10
            }
        ],
        "TPid": "CGR_API_TESTS"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested DestRateTiming profile not found.


ApierV1.GetTPDestRateTimingIds
++++++++++++++++++++++++++++

Queries DestRateTiming identities on specific tariff plan.

**Request**:

 Data:
  ::

   type AttrTPDestRateTimingIds struct {
	TPid string // Tariff plan id
   }

 Mandatory parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 12,
    "method": "ApierV1.GetTPDestRateTimingIds",
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
    "id": 12,
    "result": [
        "DRT_1CENTPERSEC"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested tariff plan not found.
