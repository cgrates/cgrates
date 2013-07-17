Apier.SetRatingProfile
++++++++++++++++++++++

Process dependencies and load a specific rating profile from storDb into dataDb.

**Request**:

 Data:
  ::

   type AttrSetRatingProfile struct {
	TPid          string
	RateProfileId string
   }

 Mandatory parameters: ``[]string{"TPid", "RateProfileId"}``

 *JSON sample*:
  ::

   {
    "id": 0, 
    "method": "Apier.SetRatingProfile", 
    "params": [
        {
            "RateProfileId": "RPF_SAMPLE_1", 
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


