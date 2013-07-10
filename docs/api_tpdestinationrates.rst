Apier.SetTPDestinationRate
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
    "id": 2, 
    "method": "Apier.SetTPDestinationRate", 
    "params": [
        {
            "DestinationRateId": "DST_RATE_1", 
            "DestinationRates": [
                {
                    "DestinationId": "FIST_DST2", 
                    "RateId": "SAMPLE_RATE_4"
                }, 
                {
                    "DestinationId": "DST_2", 
                    "RateId": "SAMPLE_RATE_4"
                }, 
                {
                    "DestinationId": "DST_3", 
                    "RateId": "SAMPLE_RATE_5"
                }
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

 ``DUPLICATE`` - The specified combination of TPid/DestinationRateId already exists in StorDb.


Apier.GetTPDestinationRate
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
    "id": 2, 
    "method": "Apier.GetTPDestinationRate", 
    "params": [
        {
            "DestinationRateId": "DST_RATE_1", 
            "TPid": "FIST_TP"
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
    "id": 2, 
    "result": {
        "DestinationRateId": "DST_RATE_1", 
        "DestinationRates": [
            {
                "DestinationId": "DST_2", 
                "RateId": "SAMPLE_RATE_4"
            }, 
            {
                "DestinationId": "DST_3", 
                "RateId": "SAMPLE_RATE_5"
            }, 
            {
                "DestinationId": "FIST_DST2", 
                "RateId": "SAMPLE_RATE_4"
            }
        ], 
        "TPid": "FIST_TP"
    }
   }


**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested DestinationRate id not found.


Apier.GetTPDestinationRateIds
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
    "id": 3, 
    "method": "Apier.GetTPDestinationRateIds", 
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
    "id": 3, 
    "result": [
        "DST_RATE_1", 
        "DST_RATE_2", 
        "DST_RATE_3"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested tariff plan not found.

