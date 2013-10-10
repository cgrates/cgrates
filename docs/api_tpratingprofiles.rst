ApierV1.SetTPRatingProfile
++++++++++++++++++++++++

Creates a new RatingProfile within a tariff plan.

**Request**:

 Data:
  ::

   type TPRatingProfile struct {
	TPid                 string             // Tariff plan id
	RatingProfileId        string             // RatingProfile id
	Tenant               string             // Tenant's Id
	TOR                  string             // TypeOfRecord
	Direction            string             // Traffic direction, *out is the only one supported for now
	Subject              string             // Rating subject, usually the same as account
	RatesFallbackSubject string             // Fallback on this subject if rates not found for destination
	RatingActivations    []RatingActivation // Activate rate profiles at specific time
   }

   type RatingActivation struct {
	ActivationTime   int64  // Time when this profile will become active, defined as unix epoch time
	DestRateTimingId string // Id of DestRateTiming profile
   }

 Mandatory parameters: ``[]string{"TPid", "RatingProfileId", "Tenant", "TOR", "Direction", "Subject", "RatingActivations"}``

 *JSON sample*:
  ::

   {
    "id": 14,
    "method": "ApierV1.SetTPRatingProfile",
    "params": [
        {
            "Direction": "*out",
            "RatesFallbackSubject": "",
            "RatingActivations": [
                {
                    "ActivationTime": "2012-01-01T00:00:00Z",
                    "DestRateTimingId": "DRT_1CENTPERSEC"
                }
            ],
            "RatingProfileId": "RP_ANY",
            "Subject": "*any",
            "TOR": "call",
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
    "id": 14, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/RatingProfileId already exists in StorDb.


ApierV1.GetTPRatingProfile
++++++++++++++++++++++++

Queries specific RatingProfile on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPRatingProfile struct {
	TPid             string // Tariff plan id
	RatingProfileId    string // RatingProfile id
   }

 Mandatory parameters: ``[]string{"TPid", "RatingProfileId"}``

 *JSON sample*:
  ::

   {
    "id": 15,
    "method": "ApierV1.GetTPRatingProfile",
    "params": [
        {
            "RatingProfileId": "RP_ANY",
            "TPid": "CGR_API_TESTS"
        }
    ]
   }
   
**Reply**:

 Data:
  ::

   type TPRatingProfile struct {
	TPid                 string             // Tariff plan id
	RatingProfileId      string             // RatingProfile id
	Tenant               string             // Tenant's Id
	TOR                  string             // TypeOfRecord
	Direction            string             // Traffic direction, *out is the only one supported for now
	Subject              string             // Rating subject, usually the same as account
	RatesFallbackSubject string             // Fallback on this subject if rates not found for destination
	RatingActivations    []RatingActivation // Activate rate profiles at specific time
   }

   type RatingActivation struct {
	ActivationTime   int64  // Time when this profile will become active, defined as unix epoch time
	DestRateTimingId string // Id of DestRateTiming profile
   }

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 15,
    "result": {
        "Direction": "*out",
        "RatesFallbackSubject": "",
        "RatingActivations": [
            {
                "ActivationTime": "2012-01-01T00:00:00Z",
                "DestRateTimingId": "DRT_1CENTPERSEC"
            }
        ],
        "RatingProfileId": "RP_ANY",
        "Subject": "*any",
        "TOR": "call",
        "TPid": "CGR_API_TESTS",
        "Tenant": "cgrates.org"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested RatingProfile profile not found.


ApierV1.GetTPRatingProfileIds
+++++++++++++++++++++++++++

Queries specific RatingProfile on tariff plan. Attribute parameters used as extra filters.

**Request**:

 Data:
  ::

   type AttrTPRatingProfileIds struct {
	TPid      string // Tariff plan id
	Tenant    string // Tenant's Id
	TOR       string // TypeOfRecord
	Direction string // Traffic direction
	Subject   string // Rating subject, usually the same as account
   }

 Mandatory parameters: ``[]string{"TPid"}``

 *JSON sample*:
  ::

   {
    "id": 16,
    "method": "ApierV1.GetTPRatingProfileIds",
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
    "id": 16,
    "result": [
        "RP_ANY"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - There is no data to be returned based on filters set.


