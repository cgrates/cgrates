.. _ResourceS:

ResourceS
=========


**ResourceS** is a standalone subsystem part of the **CGRateS** infrastructure, designed to allocate virtual resources for the generic *Events* (hashmaps) it receives.

Both receiving of *Events* as well as operational commands on the virtual resources is performed via a complete set of `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_.

Due it's real-time nature, **ResourceS** are designed towards high throughput being able to process thousands of *Events* per second. This is doable since each *Resource* is a very light object, held in memory and eventually backed up in *DataDB*.


Parameters
----------

ResourceS
^^^^^^^^^

**ResourceS** is the **CGRateS** component responsible of handling the *Resources*. 

It is configured within **resources** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

store_interval
	Time interval for backing up the stats into *DataDB*.

thresholds_conns
	Connections IDs towards *ThresholdS* component. If not defined, there will be no notifications sent to *ThresholdS* on *Resource* changes.

indexed_selects
	Enable profile matching exclusively on indexes. If not enabled, the *ResourceProfiles* are checked one by one which for a larger number can slow down the processing time. Possible values: <true|false>.

string_indexed_fields
	Query string indexes based only on these fields for faster processing. If commented out, each field from the event will be checked against indexes. If uncommented and defined as empty list, no fields will be checked.

prefix_indexed_fields
	Query prefix indexes based only on these fields for faster processing. If defined as empty list, no fields will be checked.

nested_fields
	Applied when all event fields are checked against indexes, and decides whether subfields are also checked.
	

ResourceProfile
^^^^^^^^^^^^^^^

The **ResourceProfile** is the configuration of a *Resource*. This will be performed over `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_ or *.csv* files. A profile is comprised out of the following parameters:

Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	Identifier for the *ResourceProfile*, unique within a *Tenant*.

FilterIDs
	List of *FilterProfiles* which should match in order to consider the *ResourceProfile* matching the event.

UsageTTL
	Autoexpire resource allocation after this time duration.

Limit
	The number of allocations this resource is entitled to.

AllocationMessage
	The message returned when this resource is responsible for allocation.

Blocker
	When specified, no futher resources are processed after this one.

Stored
	Enable offline backups for this resource

Weight
	Order the *Resources* matching the event. Higher value - higher priority.

ThresholdIDs
	List of ThresholdProfiles targetted by the *Resource*. If empty, the match will be done in :ref:`ThresholdS` component.


ResourceUsage
^^^^^^^^^^^^^

A **ResourceUsage** represents a counted allocation within a *Resource*. The following parameters are present within:

Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	Identifier for the *ResourceUsage*.

ExpiryTime
	Exact time when this allocation expires.

Units
	Number of units allocated by this *ResourceUsage*.


Processing logic
----------------

When a new *Event* is received, **ResourceS** will pass it to :ref:`FilterS` in order to find all *Resource* objects matching the *Event*. 

As a result of the selection process we will further get an ordered list of *Resource* which are matching the *Event* and are active at the request time. 

Depending of the *RPC API* used, we will have the following behavior further:

ResourcesForEvent
	Will simply return the list of *Resources* matching so far.

AuthorizeResources
	Out of *Resources* matching, ordered based on *Weight*, it will use the first one with available units to authorize the request. Returns *RESOURCE_UNAVAILABLE* error back in case of no available units found. No actual allocation is performed.

AllocateResource
	All of the *Resources* matching the event will be operated and requested units will be deducted, independent of being available or going on negative. The first one with value higher or equal to zero will be responsible of allocation and it's message will be returned as allocation message. If no allocation message is defined for the allocated resource, it's ID will be returned instead. 

	If no resources are allocated *RESOURCE_UNAVAILABLE* will be returned as error.

ReleaseResource
	Will release all the previously allocated resources for an *UsageID*. If *UsageID* is not found (which can be the case of restart), will perform a standard search via *FilterS* and try to dealocate the resources matching there.

Depending on configuration each *Resource* can be backed up regularly and asynchronously to DataDB so it can survive process restarts.

After each resource modification (allocation or release) the :ref:`ThresholdS` will be notified with the *Resource* itself where mechanisms like notifications or fraud-detection can be triggered.


Use cases
---------

* Monitor resources for a group of accounts(ie. based on a special field in the events).
* Limit the number of CPS for a destination/supplier/account (done via UsageTTL of 1s).
* Limit resources for a destination/supplier/account/time of day/etc.