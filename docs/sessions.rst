.. _SessionS:

SessionS
========


**SessionS** is a standalone subsystem within **CGRateS** responsible to manage virtual sessions based on events received. It is accessed via `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_.


Parameters
----------

SessionS
^^^^^^^^

Configured within **sessions** section within :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

listen_bijson
	Address where the *SessionS* listens for bidirectional JSON requests.

chargers_conns
	Connections towards :ref:`ChargerS` component to query charges for events.

rals_conns
	Connections towards :ref:`RALs` component to implement auth and balance reservation for events.

cdrs_conns
	Connections towards :ref:`CDRs` component where CDRs and session costs will be sent.

resources_conns
	Connections towards :ref:`ResourceS` component for resources management.

thresholds_conns
	Connections towards :ref:`ThresholdS` component to monitor and react to information within events.

stats_conns
	Connections towards :ref:`StatS` component to compute stat metrics for events.

suppliers_conns
	Connections towards :ref:`SupplierS` component to compute suppliers for events.

attributes_conns
	Connections towards :ref:`AttributeS` component for altering the events.

replication_conns
	Connections towards other :ref:`SessionS` components, used in case of session high-availability.

debit_interval
	Default debit interval in case of *\*prepaid* requests. Zero will disable automatic debits in favour of manual ones.

store_session_costs
	Used in case of decoupling events charging from CDR processing. The session costs debitted by *SessionS* will be stored into *StorDB.sessions_costs* table and merged into the CDR later when received.

default_usage
	Imposes the default usage for each tipe of call.

session_ttl
	Enables automatic detection/removal of stale sessions. Zero will disable the functionality.

session_ttl_max_delay
	Used in tandem with *session_ttl* to randomize disconnects in order to avoid system peaks.

session_ttl_last_used
	Used in tandem with *session_ttl* to emulate the last used information for missing terminate event.

session_ttl_usage
	Used in tandem with *session_ttl* to emulate the total usage information for the incomplete session.

session_indexes
	List of fields to index out of events. Used to speed up response time for session queries.

client_protocol
	Protocol version used when acting as a JSON-RPC client (ie: force disconnecting the sessions).

channel_sync_interval
	Sync channels at regular intervals to detect stale sessions. Zero will disable this functionality.

terminate_attempts
	Limit the number of attempts to terminate a session in case of errors.

alterable_fields
	List of fields which are allowed to be changed by update/terminate events.


Processing logic
----------------

Depends on the implementation of particular *RPC API* used.


GetActiveSessions, GetActiveSessionsCount, GetPassiveSessions, GetPassiveSessionsCount
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Returns the list of sessions based on the received filters.


SetPassiveSession
^^^^^^^^^^^^^^^^^

Used by *CGRateS* in High-Availability setups to replicate sessions between different *SessionS* nodes.


ReplicateSessions
^^^^^^^^^^^^^^^^^

Starts manually a replication process. Useful in cases when a node comes back online or entering maintenance mode.


AuthorizeEvent, AuthorizeEventWithDigest
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^


Used for event authorization. It's behaviour can be controlled via a number of different parameters:


GetAttributes
	Activates altering of the event by :ref:`AttributeS`.

AttributeIDs
	Selects only specific attribute profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

AuthorizeResources
	Activates event authorization via :ref:`ResourceS`. Returns *RESOURCE_UNAVAILABLE* if no resources left for the event.

GetMaxUsage
	Queries :ref:`RALs` for event's maximum usage allowed.

ProcessThresholds
	Sends the event to :ref:`ThresholdS` to be used in monitoring.

ThresholdIDs
	Selects only specific threshold profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

ProcessStats
	Sends the event to :ref:`StatS` for computing stat metrics.

StatIDs
	Selects only specific stat profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

GetSuppliers
	Sends the event to :ref:`SupplierS` to return the list of suppliers for it as part as authorization.

SuppliersMaxCost
	Mechanism to implement revenue assurance for suppliers coming from :ref:`SupplierS` component. Can be defined as a number or special meta variable: *\*event_cost*, assuring that the supplier cost will never be higher than event cost.

SuppliersIgnoreErrors
	Instructs to ignore suppliers with errors(ie: without price for specific destination in tariff plan). Without this setting the whole query will fail instead of just the supplier being ignored.


InitiateSession, InitiateSessionWithDigest
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Used in case of session initiation. It's behaviour can be influenced by following arguments:


GetAttributes
	Activates altering of the event by :ref:`AttributeS`.

AttributeIDs
	Selects only specific attribute profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

AllocateResources
	Process the event with :ref:`ResourceS`, allocating the matching requests. Returns *RESOURCE_UNAVAILABLE* if no resources left for the event.

InitSession
	Initiates the session executing following steps:

	* Fork session based on matched :ref:`ChargerS` profiles.

	* Start debit loops for *\*prepaid* requests if *DebitInterval* is higher than 0.

	* Index the session internally and start internal timers for detecting stale sessions.

ProcessThresholds
	Sends the event to :ref:`ThresholdS` to be used in monitoring.

ThresholdIDs
	Selects only specific threshold profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

ProcessStats
	Sends the event to :ref:`StatS` for computing stat metrics.

StatIDs
	Selects only specific stat profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.


UpdateSession
^^^^^^^^^^^^^

Used to update an existing session or initiating a new one if none found. It's behaviour can be influenced by the following arguments:

GetAttributes
	Use :ref:`AttributeS` to alter the event.

AttributeIDs
	Selects only specific attribute profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

UpdateSession
	Involves charging mechanism into processing. Following steps are further executed:

	* Relocate session if *InitialOriginID* field is present in the event.

	* Initiate session if the *CGRID* is not found within the active sessions.

	* Update timers for session stale detection mechanism.

	* Debit the session usage for all the derived *\*prepaid* sessions.


TerminateSession
^^^^^^^^^^^^^^^^

Used to terminate an existing session or to initiate+terminate a new one. It's behaviour can be influenced by the following arguments:

TerminateSession
	Stop the charging process. Involves the following steps:

	* Relocate session if *InitialOriginID* field is present in the event.

	* Initiate session if the *CGRID* is not found within the active sessions.

	* Unindex the session so it does not longer show up in active sessions queries.

	* Stop the timer for session stale detection mechanism.

	* Stop the debit loops if exist.

	* Balance the charges (refund or debit more).

	* Store the session costs if configured.

	* Cache the session for later CDRs if configured.

ReleaseResources
	Will release the aquired resources within :ref:`ResourceS`.

ProcessThresholds
	Send the event to :ref:`ThresholdS` for monitoring.

ThresholdIDs
	Selects only specific threshold profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

ProcessStats
	Send the event to :ref:`StatS` for building the stat metrics.

StatIDs
	Selects only specific stat profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.



ProcessMessage
^^^^^^^^^^^^^^

Optimized for event charging, without creating sessions based on it. Influenced by the following arguments:

GetAttributes
	Alter the event via :ref:`AttributeS`.

AttributeIDs
	Selects only specific attribute profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

AllocateResources
	Alter the event via :ref:`ResourceS` for resource allocation.

Debit
	Debit the event via :ref:`RALs`. Uses :ref:`ChargerS` to fork the event if needed.

ProcessThresholds
	Send the event to :ref:`ThresholdS` for monitoring.

ThresholdIDs
	Selects only specific threshold profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

ProcessStats
	Send the event to :ref:`StatS` for building the stat metrics.

StatIDs
	Selects only specific stat profiles (instead of discovering them via :ref:`FilterS`). Faster in processing than the discovery mechanism.

GetSuppliers
	Sends the event to :ref:`SupplierS` to return the list of suppliers for it.

SuppliersMaxCost
	Mechanism to implement revenue assurance for suppliers coming from :ref:`SupplierS` component. Can be a number or special meta variable: *\*event_cost*, assuring that the supplier cost will never be higher than event cost.

SuppliersIgnoreErrors
	Instructs to ignore suppliers with errors(ie: without price for specific destination in tariff plan). Without this setting the whole query will fail instead of just the supplier being ignored.



ProcessCDR
^^^^^^^^^^

Build the CDR out of the event and send it to :ref:`CDRs`. It has the ability to use cached sessions for obtaining additional information like fields with values or derived charges, forking also the CDR based on that.


ProcessEvent
^^^^^^^^^^^^

Will generically process an event, having the ability to merge all the functionality of previous processing APIs. 

Instead of arguments, the options for enabling various functionaity will come in the form of *Flags*. These will be of two types: **main** and **auxiliary**, the last ones being considered suboptions of the first. The available flags are:


\*attributes
	Activates altering of the event via :ref:`AttributeS`.

\*cost
	Queries :ref:`RALs` for event cost.

\*resources
	Process the event with :ref:`ResourceS`. Additional auxiliary flags can be specified here:

	**\*authorize**
		Authorize the event.

	**\*allocate**
		Allocate resources for the event.

	**\*release**
		Release the resources used for the event.

\*rals
	Process the event with :ref:`RALs`. Auxiliary flags available:

	**\*authorize**
		Authorize the event.

	**\*initiate**
		Initialize a session out of event.

	**\*update**
		Update a sesssion (or initialize + update) out of event.

	**\*terminate**
		Terminate a session (or initialize + terminate) out of event.

\*suppliers
	Process the event with :ref:`Suppliers`. Auxiliary flags available:

	**\*ignore_errors**
		Ignore the suppliers with errors instead of failing the request completely.

	**\*event_cost**
		Ignore suppliers with cost higher than the event cost.

\*thresholds
	Process the event with :ref:`ThresholdS` for monitoring.

\*stats
	Process the event with :ref:`StatS` for metrics calculation.

\*cdrs
	Create a CDR out of the event with :ref:`CDRs`.


GetCost
^^^^^^^

Queries the cost for event from :ref:`RALs`. Additional processing options can be selected via the *Flags* argument. Possible flags:

\*attributes
	Use :ref:`AttributeS` to alter the event before cost being calculated.


SyncSessions
^^^^^^^^^^^^

Manually initiate a sync sessions mechanism. All the connections will be synced and stale sessions will be automatically disconnected.


ForceDisconnect
^^^^^^^^^^^^^^^

Disconnect the session matching the filter.


ActivateSessions
^^^^^^^^^^^^^^^^

Manually activate a session which is marked as passive.


DeactivateSessions
^^^^^^^^^^^^^^^^^^

Manually deactivate a session which is marked as active.