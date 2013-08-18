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
    "id": 0, 
    "method": "ApierV1.SetRatingProfile", 
    "params": [
        {
            "RatingProfileId": "RPF_SAMPLE_1", 
            "TPid": "TPID_SAMPLE_1"
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


