.. _ThresholdS:

ThresholdS
==========


**ThresholdS** is a standalone subsystem within **CGRateS** responsible to execute a list of *Actions* for a specific event received to process. It is accessed via `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_.

As most of the other subsystems, it is performance oriented, stored inside *DataDB* but cached inside the *cgr-engine* process. 
Caching can be done dynamically/on-demand or at start-time/precached and it is configurable within *cache* section in the :ref:`JSON configuration <configuration>`.


Processing logic
----------------

When a new *Event* is received, **ThresholdS** will pass it to :ref:`FilterS` in order to find all *SupplierProfiles* matching the *Event*. 

As a result of the selection process we will get a list of :ref:`Thresholds<Threshold>` matching the *Event* and are active at the *EventTime*. 



APIs logic
----------


GetThresholdIDs
^^^^^^^^^^^^^^^

Returns a list of *ThresholdIDs* configured on a *Tenant*.


GetThresholdsForEvent
^^^^^^^^^^^^^^^^^^^^^

Returns a list of :ref:`Thresholds<Threshold>` matching the event.


GetThreshold
^^^^^^^^^^^^

Returns a specific :ref:`Threshold` based on it's *Tenant* and *ID*.


ProcessEvent
^^^^^^^^^^^^

Technically processes the *Event*, executing all the *Actions* configured within all the matching :ref:`Thresholds<Threshold>`.


Parameters
----------


ThresholdS
^^^^^^^^^^

**ThresholdS** is the **CGRateS** component responsible for handling the :ref:`Thresholds<Threshold>`.

It is configured within **thresholds** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

store_interval
	Time interval for backing up the thresholds into *DataDB*.

indexed_selects
	Enable profile matching exclusively on indexes. If not enabled, the :ref:`Thresholds<Threshold>` are checked one by one which for a larger number can slow down the processing time. Possible values: <true|false>.

string_indexed_fields
	Query string indexes based only on these fields for faster processing. If commented out, each field from the event will be checked against indexes. If defined as empty list, no fields will be checked.

prefix_indexed_fields
	Query prefix indexes based only on these fields for faster processing. If defined as empty list, no fields will be checked.

nested_fields
	Applied when all event fields are checked against indexes, and decides whether subfields are also checked.


.. _ThresholdProfile:

ThresholdProfile
^^^^^^^^^^^^^^^^

Contains the configuration to create a :ref:`Threshold`. Following fields can be defined:

Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	The profile identificator.

FilterIDs
	List of *FilterProfileIDs* which should match in order to consider the profile matching the event.

MaxHits
	Limit number of hits for this threshold. Once this is reached, the threshold is considered disabled.

MinHits
	Only execute actions after this number is reached.

MinSleep
	Disable the threshold for consecutive hits for the duration of *MinSleep*.

Blocker
	Do not process thresholds who's *Weight* is lower.

Weight
	Sorts the execution of multiple thresholds matching the event. The higher the *Weight* is, the higher the priority to be executed.

ActionProfileIDs
	List of *ActionProfiles* to execute for this threshold.

Async
	If true, do not wait for actions to complete.


.. _Threshold:

Threshold
^^^^^^^^^

Represents one threshold, instantiated from a :ref:`ThresholdProfile`. It contains the following fields:


Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	The threshold identificator.

Hits
	Number of hits so far.

Snooze
	If initialized, it will contain the time when this threshold will become active again.



Use cases
---------

* Improve network transparency and automatic reaction to outages monitoring stats produced by :ref:`StatS`.
* Monitor active channels used by a supplier/customer/reseller/destination/weekends/etc out of :ref:`ResourceS` events.
* Monitor balance consumption out of *Account* events.
* Monitor calls out of :ref:`CDRs` events or :ref:`SessionS`.
* Fraud detection with automatic mitigation based of all events mentioned above.