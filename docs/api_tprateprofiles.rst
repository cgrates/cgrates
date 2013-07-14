Apier.SetTPRateProfile
++++++++++++++++++++++

Creates a new RateProfile within a tariff plan.

**Request**:

 Data:
  ::

   type TPRateProfile struct {
	TPid                 string             // Tariff plan id
	RateProfileId        string             // RateProfile id
	Tenant               string             // Tenant's Id
	TOR                  string             // TypeOfRecord
	Direction            string             // Traffic direction, OUT is the only one supported for now
	Subject              string             // Rating subject, usually the same as account
	RatesFallbackSubject string             // Fallback on this subject if rates not found for destination
	RatingActivations    []RatingActivation // Activate rate profiles at specific time
   }

   type RatingActivation struct {
	ActivationTime   int64  // Time when this profile will become active, defined as unix epoch time
	DestRateTimingId string // Id of DestRateTiming profile
   }

 Mandatory parameters: ``[]string{"TPid", "RateProfileId", "Tenant", "TOR", "Direction", "Subject", "RatingActivations"}``

 *JSON sample*:
  ::

   {
    "id": 3, 
    "method": "Apier.SetTPRateProfile", 
    "params": [
        {
            "Direction": "OUT", 
            "RateProfileId": "SAMPLE_RP_2", 
            "RatingActivations": [
                {
                    "ActivationTime": 1373609003, 
                    "DestRateTimingId": "DSTRTTIME_1"
                }, 
                {
                    "ActivationTime": 1373609004, 
                    "DestRateTimingId": "DSTRTTIME_2"
                }
            ], 
            "Subject": "dan", 
            "TOR": "CALL", 
            "TPid": "SAMPLE_TP", 
            "Tenant": "Tenant1"
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
    "id": 3, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``DUPLICATE`` - The specified combination of TPid/RateProfileId already exists in StorDb.


Apier.GetTPRateProfile
++++++++++++++++++++++

Queries specific RateProfile on tariff plan.

**Request**:

 Data:
  ::

   type AttrGetTPRateProfile struct {
	TPid             string // Tariff plan id
	RateProfileId    string // RateProfile id
   }

 Mandatory parameters: ``[]string{"TPid", "RateProfileId"}``

 *JSON sample*:
  ::

   {
    "id": 0, 
    "method": "Apier.GetTPRateProfile", 
    "params": [
        {
            "RateProfileId": "SAMPLE_RP_2", 
            "TPid": "SAMPLE_TP"
        }
    ]
   }
   
**Reply**:

 Data:
  ::

   type TPRateProfile struct {
	TPid                 string             // Tariff plan id
	RateProfileId        string             // RateProfile id
	Tenant               string             // Tenant's Id
	TOR                  string             // TypeOfRecord
	Direction            string             // Traffic direction, OUT is the only one supported for now
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
    "id": 0, 
    "result": {
        "Direction": "OUT", 
        "RateProfileId": "SAMPLE_RP_2", 
        "RatesFallbackSubject": "", 
        "RatingActivations": [
            {
                "ActivationTime": 1373609003, 
                "DestRateTimingId": "DSTRTTIME_1"
            }, 
            {
                "ActivationTime": 1373609004, 
                "DestRateTimingId": "DSTRTTIME_2"
            }
        ], 
        "Subject": "dan", 
        "TOR": "CALL", 
        "TPid": "SAMPLE_TP", 
        "Tenant": "Tenant1"
    }
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - Requested RateProfile profile not found.


Apier.GetTPRateProfileIds
+++++++++++++++++++++++++

Queries specific RateProfile on tariff plan. Attribute parameters used as extra filters.

**Request**:

 Data:
  ::

   type AttrTPRateProfileIds struct {
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
    "id": 0, 
    "method": "Apier.GetTPRateProfileIds", 
    "params": [
        {
            "Subject": "dan", 
            "TPid": "SAMPLE_TP", 
            "Tenant": "Tenant1"
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
    "id": 0, 
    "result": [
        "SAMPLE_RP_1", 
        "SAMPLE_RP_2"
    ]
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - There is no data to be returned based on filters set.


