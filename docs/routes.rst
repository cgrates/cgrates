.. _Asterisk: https://www.asterisk.org/
.. _FreeSWITCH: https://freeswitch.com/
.. _Kamailio: https://www.kamailio.org/w/
.. _OpenSIPS: https://opensips.org/


.. _Routes:

RouteS
=========


**RouteS** is a standalone subsystem within **CGRateS** responsible to compute a list of routes which can be used for a specific event received to process. It is accessed via `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_.

As most of the other subsystems, it is performance oriented, stored inside *DataDB* but cached inside the *cgr-engine* process. 
Caching can be done dynamically/on-demand or at start-time/precached and it is configurable within *cache* section in the :ref:`JSON configuration <configuration>`.


Processing logic
----------------

When a new *Event* is received, **RouteS** will pass it to :ref:`FilterS` in order to find all :ref:`SupplierProfiles<SupplierProfile>` matching the *Event*. 

As a result of the selection process we will get a single :ref:`SupplierProfile` matching the *Event*, is active at the *EventTime* and has a higher priority than the other matching :ref:`SupplierProfiles<SupplierProfile>`. 

Depending on the *Strategy* defined in the *SupplierProfile*, further steps will be taken (ie: query cost, stats, ordering) for each of the individual *SupplierIDs* defined within the *SupplierProfile*.


APIs logic
----------

GetSupplierProfilesForEvent
^^^^^^^^^^^^^^^^^^^^^^^^^^^

Given the *Event* it will return a list of ordered *SupplierProfiles* matching at the *EventTime*. 

This API is useful to test configurations.


GetRoutes
^^^^^^^^^^^^

Will return a list of *Routes* from within a *SupplierProfile* ordered based on *Strategy*.


Parameters
----------


RouteS
^^^^^^^^^

**RouteS** is the **CGRateS** component responsible for handling the *SupplierProfiles*.

It is configured within **routes** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

indexed_selects
	Enable profile matching exclusively on indexes. If not enabled, the *SupplierProfiles* are checked one by one which for a larger number can slow down the processing time. Possible values: <true|false>.

string_indexed_fields
	Query string indexes based only on these fields for faster processing. If commented out, each field from the event will be checked against indexes. If defined as empty list, no fields will be checked.

prefix_indexed_fields
	Query prefix indexes based only on these fields for faster processing. If defined as empty list, no fields will be checked.

nested_fields
	Applied when all event fields are checked against indexes, and decides whether subfields are also checked.

attributes_conns
	Connections to AttributeS for altering events before supplier queries. If undefined, fields modifications are disabled.

resources_conns
	Connections to ResourceS for *res sorting, empty to disable functionality.

stats_conns
	Connections to StatS for *stats sorting, empty to disable stats functionality.

default_ratio
	Default ratio used in case of *load strategy


.. _SupplierProfile:

SupplierProfile
^^^^^^^^^^^^^^^

Contains the configuration for a set of routes which will be returned in case of match. Following fields can be defined:

Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	The profile identificator.

FilterIDs
	List of *FilterProfileIDs* which should match in order to consider the profile matching the event.

Sorting
	Sorting strategy applied when ordering the individual *Routes* defined bellow. Possible values are:

	**\*weight**
		Classic method of statically sorting the routes based on their priority.

	**\*lc**
		LeastCost will sort the routes based on their cost (lowest cost will have higher priority). If two routes will be identical as cost, their *Weight* will influence the sorting further. If *AccountIDs* will be specified, bundles can be also used during cost calculation, the only condition is that the bundles should cover complete usage.

		The following fields are mandatory for cost calculation: *Account*/*Subject*, *Destination*, *SetupTime*. *Usage* is optional and if present in event, it will be used for the cost calculation.

	**\*hc**
		HighestCost will sort the routes based on their cost(higher cost will have higher priority). If two routes will be identical as cost, their *Weight* will influence the sorting further.

		The following fields are mandatory for cost calculation: *Account*/*Subject*, *Destination*, *SetupTime*. *Usage* is optional and if present in event, it will be used for the cost calculation.

	**\*qos**
		QualityOfService strategy will sort the routes based on their stats. It takes the StatIDs to check from the supplier *StatIDs* definition. The metrics used as part of sorting are to be defined in *SortingParameters* field bellow. If Stats are missing the metrics defined in *SortingParameters* defaults for those will be populated for order (10000000 as PDD and -1 for the rest).

	**\*reas**
		ResourceAscendentSorter will sort the routes based on their resource usage, lowest usage giving higher priority. The resources will be queried for each supplier based on it's *ResourceIDs* field and the final usage for each supplier will be given by the sum of all the resource usages queried.

	**\*reds**
		ResourceDescendentSorter will sort the routes based on their resource usage, highest usage giving higher priority. The resources will be queried for each supplier based on it's *ResourceIDs* field and the final usage for each supplier will be given by the sum of all the resource usages queried.

	**\*load**
		LoadDistribution will sort the routes based on their load. An important parameter is the *\*ratio* which is defined as *supplierID:Ratio* within the SortingParameters. If no supplierID is present within SortingParameters, the system will look for *\*default* or fallback in the configuration to *default_ratio* within :ref:`JSON configuration <configuration>`. The *\*ratio* will specify the probability to get traffic on a *Supplier*, the higher the *\*ratio* more chances will a *Supplier* get for traffic. 

		The load will be calculated out of the *StatIDs* parameter of each *Supplier*. It is possible to also specify there directly the metric being used in the format *StatID:MetricID*. If only *StatID* is instead specified, all metrics will be summed to get the final value. 


SortingParameters
	Will define additional parameters for each strategy. Following extra parameters are available(based on strategy):

	**\*qos**
		List of metrics to be used for sorting in order of importance.

Weight
	Priority in case of multiple *SupplierProfiles* matching an *Event*. Higher *Weight* will have more priority.

Routes
	List of :ref:`Supplier` objects which are part of this *SupplierProfile*


.. _Supplier:

Supplier
^^^^^^^^

The *Supplier* represents one supplier within the *SupplierProfile*. Following parameters are defined for it:

ID
	Supplier ID, will be returned via APIs. Should be known on the remote side and match some business logic (ie: gateway id or directly an IP address).

FilterIDs
	List of *FilterProfileIDs* which should match in order to consider the *Supplier* in use/active.
	
AccountIDs
	List of account IDs which should be checked in case of some strategies (ie: *lc, *hc).
	
RatingPlanIDs
	List of RatingPlanIDs which should be checked in case of some strategies (ie: *lc, *hc).

ResourceIDs
	List of ResourceIDs which should be checked in case of some strategies (ie: *reas or *reds).

StatIDs
	List of StatIDs which should be checked in case of some strategies (ie: *qos or *load). Can also be defined as *StatID:MetricID*.

Weight
	Used for sorting in some strategies (ie: *weight, *lc or *hc).

Blocker
	No more routes are provided after this one.
	
SupplierParameters
	Container which is trasparently passed to the remote client to be used in it's own logic (ie: gateway prefix stripping or other gateway parameters).



Use cases
---------

* Calculate LCR directly by querying APIs (GetRoutes).
* LCR system together with  Kamailio_ *dispatcher* module where the *SupplierID* whithin *CGRateS* will be used as dispatcher set within Kamailio_.
* LCR system together with OpenSIPS_ drouting module where the *SupplierID* whithin *CGRateS* will be used as drouting carrier id.
* LCR system together with FreeSWITCH_ or Asterisk_ where the *SupplierID* whithin *CGRateS* will be used as gateway ID within the dialplan of FreesWITCH_ or Asterisk_.