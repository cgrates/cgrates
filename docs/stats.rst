StatS
=====


**StatS** is a standalone subsystem part of the **CGRateS** infrastructure, designed to aggregate and calculate statistical metrics for the generic *Events* (hashmaps) it receives.

Both receiving of *Events* as well as *Metrics* displaying is performed via a complete set of :ref:`CGRateS RPC APIs<remote-management>`.

Due it's real-time nature, **StatS** are designed towards high throughput being able to process thousands of *Events* per second. This is doable since each *StatQueue* is a very light object, held in memory and eventually backed up in *DataDB*.

Processing logic
----------------

When a new *Event* is received, **StatS** will pass it to :ref:`FilterS` in order to find all *StatProfiles* matching the *Event*. 

As a result of the selection process we will further get an ordered list of *StatProfiles* which are matching the *Event* and are active at the request time. 

For each of these profiles we will further calculate the metrics it has configured for the *Event* received. If *ThresholdIDs* are not *\*none*, we will include the *Metrics* into special *StatUpdate* events, defined internally, and pass them further to the [ThresholdS](ThresholdS) for processing.

Depending on configuration each *StatQueue* can be backed up regularly and asynchronously to DataDB so it can survive process restarts.


Parameters
----------


StatQueueProfile
^^^^^^^^^^^^^^^^

Ís made of the following fields:

Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	Identifier for the *StatQueueProfile*, unique within a *Tenant*.

FilterIDs
	List of *FilterProfiles* which should match in order to consider the *StatQueueProfile* matching the event.

ActivationInterval
	The time interval when this profile becomes active. If undefined, the profile is always active. Other options are start time, end time or both.

QueueLength
	Maximum number of items stored in the queue. Once the *QueueLength* is reached, new items entering will cause oldest one to be dropped (FIFO mode).

TTL
	Time duration causing items in the queue to expire and be removed automatically from the queue.

Metrics
	List of statistical metrics to build for items within this *StatQueue*. See [bellow](#statqueue-metrics) for possible values here.

ThresholdIDs
	List of threshold IDs to check on when new items are updating the queue metrics.

Blocker
	Do not process further *StatQueues*.

Stored
	Enable offline backups for this *StatQueue*

Weight
	Order the *StatQueues* matching the event. Higher value - higher priority.

MinItems
	Display metrics only if the number of items in the queue is higher than this.


StatQueue Metrics
^^^^^^^^^^^^^^^^^

Following metrics are implemented:

\*asr
	`Answer-seizure ratio <https://en.wikipedia.org/wiki/Answer-seizure_ratio>`_. Relies on *AnswerTime* field in the *Event*.
\*acd
	`Average call duration <https://en.wikipedia.org/wiki/Average_call_duration>`. Uses *AnswerTime* and *Usage* fields in the *Event*.
\*tcd
	Total call duration. Uses *Usage* out of *Event*.

\*acc
	Average call cost. Uses *Cost* field out of *Event*.

\*tcc
	Total call cost. Uses *Cost* field out of *Event*.

\*pdd
	`Post dial delay <https://www.voip-info.org/pdd/>`. Uses *PDD* field in the event.

\*ddc
	Distinct destination count will keep the number of unique destinations found in *Events*. Relies on *Destination* field in the *Event*.

\*sum
	Generic metric to calculate mathematical sum for a specific field in the *Events*. Format: <*\*sum#FieldName*>.

\*average
	Generic metric to calculate the mathematical average of a specific field in the *Events*. Format: <*\*average#FieldName*>.

\*distinct
	Generic metric to return the distinct number of appearance of a field name within *Events*. Format: <*\*distinct#FieldName*>.


Use cases
---------

* Aggregate various traffic metrics for traffic transparency.
* Revenue assurance applications.
* Fraud detection by aggregating specific billing metrics during sensitive time intervals (*acc, *tcc, *tcd).
* Building call patterns.
* Building statistical information to train systems capable of artificial intelligence.
* Building quality metrics used in traffic routing.
