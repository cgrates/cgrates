Apier.SetTPDestRateTiming
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
    "id": 0, 
    "method": "Apier.SetTPDestRateTiming", 
    "params": [
        {
            "DestRateTimingId": "SAMPLE_DRTIMING_1", 
            "DestRateTimings": [
                {
                    "DestRatesId": "SAMPLE_DR_1", 
                    "TimingId": "SAMPLE_TIMING_1", 
                    "Weight": 10
                }
            ], 
            "TPid": "SAMPLE_TP"
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

 ``DUPLICATE`` - The specified combination of TPid/DestRateTimingId already exists in StorDb.


Apier.GetTPDestRateTiming
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
    "id": 4, 
    "method": "Apier.GetTPDestRateTiming", 
    "params": [
        {
            "DestRateTimingId": "SAMPLE_DRTIMING_1", 
            "TPid": "SAMPLE_TP"
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
    "id": 4, 
    "result": {
        "DestRateTimingId": "SAMPLE_DRTIMING_1", 
        "DestRateTimings": [
            {
                "DestRatesId": "SAMPLE_DR_1", 
                "TimingId": "SAMPLE_TIMING_1", 
                "Weight": 10
            }
        ], 
        "TPid": "SAMPLE_TP"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested DestRateTiming profile not found.


Apier.GetTPDestRateTimingIds
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
    "id": 5, 
    "method": "Apier.GetTPDestRateTimingIds", 
    "params": [
        {
            "TPid": "SAMPLE_TP"
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
        "SAMPLE_DRTIMING_1", 
        "SAMPLE_DRTIMING_2", 
        "SAMPLE_DRTIMING_3"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested tariff plan not found.
