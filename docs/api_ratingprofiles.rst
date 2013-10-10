ApierV1.SetRatingProfile
++++++++++++++++++++++

Process dependencies and load a specific rating profile from storDb into dataDb.

**Request**:

 Data:
  ::

   type AttrSetRatingProfile struct {
	TPid          string
	RatingProfileId string
   }

 Mandatory parameters: ``[]string{"TPid", "RatingProfileId"}``

 *JSON sample*:
  ::

   {
    "id": 37,
    "method": "ApierV1.SetRatingProfile",
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

   string

 Possible answers:
  ``OK`` - Success.

 *JSON sample*:
  ::

   {
    "error": null, 
    "id": 37, 
    "result": "OK"
   }

**Errors**:

 ``MANDATORY_IE_MISSING`` - Mandatory parameter missing from request.

 ``SERVER_ERROR`` - Server error occurred.


