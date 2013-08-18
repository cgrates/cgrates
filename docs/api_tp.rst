ApierV1.GetTPIds
+++++++++++++++++++++++++

Queries tarrif plan identities gathered from all tables.

**Request**:

 Data:
  ::

   type AttrGetTPIds struct {
   }

 *JSON sample*:
  ::

   {
    "id": 9, 
    "method": "ApierV1.GetTPIds", 
    "params": []
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
        "SAMPLE_TP", 
        "SAMPLE_TP_2"
    ]
   }



**Errors**:

 ``SERVER_ERROR`` - Server error occurred.

 ``NOT_FOUND`` - No tariff plans defined.
