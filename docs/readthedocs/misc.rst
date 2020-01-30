8. GetCacheStats
================

GetCacheStats returns datadb cache status. Empty params return all stats:

:Hint:

    cgr> cache_stats

*Request*

::

   {
   	"method": "APIerSv1.GetCacheStats",
   	"params": [{}],
   	"id": 0
   }

*Response:*

::

   {
   	"id": 0,
   	"result": {
   		"Destinations": 0,
   		"ReverseDestinations": 0,
   		"RatingPlans": 4,
   		"RatingProfiles": 0,
   		"Actions": 0,
   		"ActionPlans": 4,
   		"AccountActionPlans": 0,
   		"SharedGroups": 0,
   		"DerivedChargers": 0,
   		"LcrProfiles": 0,
   		"CdrStats": 6,
   		"Users": 3,
   		"ResourceProfiles": 0,
   		"Resources": 0,
   		"StatQueues": 0,
   		"StatQueueProfiles": 0,
   		"Thresholds": 0,
   		"ThresholdProfiles": 0,
   		"Filters": 0
   	},
   	"error": null
   }
