.. _stats:

StatS
=====


**StatS** is a standalone subsystem part of the **CGRateS** infrastructure, designed to aggregate and calculate statistical metrics for the generic *Events* (hashmaps) it receives.

Both receiving of *Events* as well as *Metrics* displaying is performed via a complete set of `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_.

Due it's real-time nature, **StatS** are designed towards high throughput being able to process thousands of *Events* per second. This is doable since each *StatQueue* is a very light object, held in memory and eventually backed up in *DataDB*.


Processing logic
----------------

When a new *Event* is received, **StatS** will pass it to :ref:`FilterS` in order to find all *StatProfiles* matching the *Event*. 

As a result of the selection process we will further get an ordered list of *StatProfiles* which are matching the *Event* and are active at the request time. 

For each of these profiles we will further calculate the metrics it has configured for the *Event* received. If *ThresholdIDs* are not *\*none*, we will include the *Metrics* into special *StatUpdate* events, defined internally, and pass them further to the [ThresholdS](ThresholdS) for processing.

Depending on configuration each *StatQueue* can be backed up regularly and asynchronously to DataDB so it can survive process restarts.


Parameters
----------


StatS
^^^^^

**StatS** is the **CGRateS** component responsible of handling the *StatQueues*. 

It is configured within **stats** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

store_interval
	Time interval for backing up the stats into *DataDB*.

store_uncompressed_limit
	After this limit is hit the events within *StatQueue* will be stored aggregated.

thresholds_conns
	Connections IDs towards *ThresholdS* component. If not defined, there will be no notifications sent to *ThresholdS* on *StatQueue* changes.

indexed_selects
	Enable profile matching exclusively on indexes. If not enabled, the *StatQueues* are checked one by one which for a larger number can slow down the processing time. Possible values: <true|false>.

string_indexed_fields
	Query string indexes based only on these fields for faster processing. If commented out, each field from the event will be checked against indexes. If uncommented and defined as empty list, no fields will be checked.

prefix_indexed_fields
	Query prefix indexes based only on these fields for faster processing. If defined as empty list, no fields will be checked.

nested_fields
	Applied when all event fields are checked against indexes, and decides whether subfields are also checked.


StatQueueProfile
^^^^^^^^^^^^^^^^

Ís made of the following fields:

Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	Identifier for the *StatQueueProfile*, unique within a *Tenant*.

FilterIDs
	List of *FilterProfileIDs* which should match in order to consider the profile matching the event.

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
	`Average call duration <https://en.wikipedia.org/wiki/Average_call_duration>`_. Uses *AnswerTime* and *Usage* fields in the *Event*.
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
