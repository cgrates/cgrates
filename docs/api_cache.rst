Cache APIs
==========

Set of cache related APIs.


ApierV1.ReloadCache
-------------------

Used to enforce a cache reload. It can be fine tuned to reload individual destinations and rating plans. In order to reload all destinations and/or rating plans, one can use empty list or null values instead.

**Request**:

Data:

 ::

  type ApiReloadCache struct {
	DestinationIds      []string
	RatingPlanIds       []string
   }

 Mandatory parameters: none

 *JSON sample*:
  ::

   {
    "id": 1,
    "method": "ApierV1.ReloadCache",
    "params": [
        {
            "DestinationIds": [
                "GERMANY",
                "GERMANY_MOBILE",
                "FS_USERS"
            ],
            "RatingPlanIds": [
                "RETAIL1"
            ]
        }
    ]
   }

**Reply**:

 Data:
  ::

   string

 Possible answers:
   * *OK*

 *JSON sample*:
  ::

   {
    "error": null,
    "id": 1,
    "result": "OK"
   }

**Errors**:

 ``SERVER_ERROR`` - Server error occurred.
