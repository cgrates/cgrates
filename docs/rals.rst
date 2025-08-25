.. _rals:

RALs
====


**RALs** is a standalone subsystem within **CGRateS** designed to handle two major tasks: :ref:`Rating` and :ref:`Accounting`. It is accessed via `CGRateS RPC APIs <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_.



.. _Rating:

Rating
------

Rating is the process responsible to attach costs to events.

The costs are calculated based on the input data defined within :ref:`TariffPlan` in the following sections:


.. _RatingProfile:

RatingProfile
^^^^^^^^^^^^^

Binds the event via a fixed number of fields to a predefined :ref:`RatingPlan`. Configured via the following parameters:

Tenant
	The tenant on the platform (one can see the tenant as partition ID). Matched from event or inherited from :ref:`JSON configuration <configuration>`.

Category
	Freeform field used to "categorize" the event. Matched from event or inherited from :ref:`JSON configuration <configuration>`.

Subject
	Rating subject matched from the event. In most of the cases this equals with the *Account* using the service.

ActivationTime
	Date and time when the profile becomes active. There is no match before this date.

RatingPlanID
	Identifier of the :ref:`RatingPlan` assigned to the event.

FallbackSubjects
	List of rating subjects which will be searched in order in case of missing rates in case of defined :ref:`RatingPlan`. This list is only considered at first level of iteration (not considering *FallbackSubjects* within interations).

.. Note:: One *RatingProfile* entry is composed out of a unique combination of *Tenant* + *Category* + *Subject*.


.. _RatingPlan:

RatingPlan
^^^^^^^^^^

Groups together rates per destination and relates them to event timing. Configured via the following parameters:

ID
	The tag uniquely idenfying each RatingPlan. There can be multiple entries grouped by the same ID.

DestinationRatesID
	The identifier of the :ref:`DestinationRate` set.

TimingID
	The itentifier of the :ref:`Timing` profile.

Weight
	Priority of matching rule (*DestinationRatesID*+*TimingID*). Higher value equals higher priority.


.. _DestinationRate:

DestinationRate
^^^^^^^^^^^^^^^

Groups together destination with rate profiles and assigns them some properties used in the rating process. Configured via the following parameters:

ID
	The tag uniquely idenfying each DestinationRate profile. There can be multiple entries grouped by the same ID.

DestinationsID
	The identifier of the :ref:`Destination` profile.

RatesID
	The identifier of the :ref:`Rate` profile.

RoundingMethod
	Method used to round during float operations. Possible values:

	**\*up**
		Upsize towards next integer value (ie: 0.11 -> 0.2)

	**\*middle**
		Round at middle towards next integer value (ie: 0.11 -> 0.1, 0.16 -> 0.2)

	**\*down**
		Downsize towards next integer (ie: 0.19 -> 0.1).

RoundingDecimals
	Number of decimals after the comma to use when rounding floats.

MaxCost
	Maximum cost threshold for an event or session.

MaxCostStrategy
	The strategy used once the maximum cost is reached. Can be one of following options:

	**\*free**
		Anything above *MaxCost* is not charged

	**\*disconnect**
		The session is disconnected forcefully. 


.. _Destination:

Destination
^^^^^^^^^^^

Groups list of prefixes under one *Destination* profile. Configured via the following parameters:

ID
	The tag uniquely idenfying each Destination profile. There can be multiple entries grouped by the same ID.

Prefix
	One prefix entry (can be also full destination string).


.. _Rate:

Rate
^^^^

A *Rate* profile will contain all the individual rates applied for a matching event/session on a time interval. Configured via the following parameters:

ID
	The tag uniquely idenfying each *Rate* profile. There can be multiple entries grouped by the same ID.

ConnectFee
	One time charge applying when the session is opened.

Rate
	The rate applied for one rating increment.

RateUnit
	The unit raported to the usage received.

RateIncrement
	Splits the usage received into smaller increments.

GroupIntervalStart
	Activates the rate at specific usage within the event.


.. _Timing:

Timing
^^^^^^

A *Timing* profile is giving time awarness to an event. Configured via the following parameters:

ID
	The tag uniquely idenfying each *Timing* profile.

Years
	List of years to match within the event. Defaults to the catch-all meta: *\*any*.

Months
	List of months to match within the event. Defaults to the catch-all meta: *\*any*.

MonthDays
	List of month days to match within the event. Defaults to the catch-all meta: *\*any*.

WeekDays
	List of week days to match within the event as integer values. Special case for *Sunday* which matches for both 0 and 7.

Time
	The exact time to match (mostly as time start). Defined in the format: *hh:mm:ss*



.. Note:: Due to optimization, CGRateS encapsulates and stores the rating information into just three objects: *Destinations*, *RatingProfiles* and *RatingPlan* (composed out of *RatingPlan*, *DestinationRate*, *Rate* and *Timing* objects).



.. _Accounting:

Accounting
----------

Accounting is the process of charging an *Account* on it's *Balances*. The amount of charges is decided by either internal configuration of each *Balance* or calculated by :ref:`Rating`.


.. _Account:

Account
^^^^^^^

Is the central unit of the :ref:`Accounting`. It contains the following fields:


Tenant
	The tenant to whom the account belogs.

ID
	The Account identifier which should be unique within a tenant. This should match with the event's *Account* field.

BalanceMap
	The pool of :ref:`Balances <Balance>` indexed by type.

UnitCounters
	Usage counters which are set out of thresholds defined in :ref:`ActionTriggers <ActionTrigger>`

AllowNegative
	Allows authorization independent on credit available.

UpdateTime
	Set on each update in DataDB.

Disabled
	Marks the account as disabled, making it invisible to charging.



.. _Balance:

Balance
^^^^^^^


Is the unit container (wallet/bundle) of the :ref:`Account`. There can be unlimited number of *Balances* within one :ref:`Account`, groupped by their type.

The following *BalanceTypes* are supported:

\*voice
	Coupled with voice calls, represents nanosecond units.

\*data
	Coupled with data sessions, represents units of data (virtual units).

\*sms
	Coupled with SMS events, represents number of SMS units.

\*mms
	Coupled with MMS events, represents number of MMS units.

\*generic
	Matching all types of events after specific ones, representing generic units (i.e., for each x \*voice minutes, y \*sms units, and z \*data units will have their respective usage)

\*monetary
	Matching all types of events after specific ones, representing monetary units (can be interpreted as virtual currency).



A *Balance* is made of the following fields:

Uuid
	Unique identifier within the system (unique hash generated for each *Balance*).

ID
	Idendificator configurable by the administrator. It is unique within an :ref:`Account`.

Value
	The *Balance's* value.

ExpirationDate
	The expiration time of this *Balance*

Weight
	Used to prioritize matching balances for an event. The higher the *Weight*, the more priority for the *Balance*.

DestinationIDs
	List of :ref:`Destination` profiles this *Balance* will match for, considering event's *Destination* field.

RatingSubject
	The rating subject this balance will use when calculating the cost. 

	This will match within :ref:`RatingProfile`.  If the rating profile starts with character *\**, special cost will apply, without interogating :ref:`Rating` for it. The following *metas* are available:

	**\*zero$xdur**
		A *\*zero* followed by a duration will be the equivalent of 0 cost, charged in increments of *x* duration (ie: *\*zero1m*.

	**\*any**
		Points out to default (same as undefined). Defaults are set to *\*zero1s* for voice and *\*zero1ns* for everything else.

Categories
	List of event *Category* fields this *Balance* will match for.

SharedGroup
	Pointing towards a shared balance ID.

TimingIDs
	List of :ref:`Timing` profiles this *Balance* will match for, considering event's *AnswerTime* field.

Disabled
	Makes the *Balance* invisible to charging.

Factors
	Used in case of of *\*generic* *BalanceType* to specify the conversion factors for different type of events.

Blocker
	A *blocking Balance* will prevent processing further matching balances when empty.



.. _ActionTrigger:

ActionTrigger
-------------

Is a mechanism to monitor Balance values during live operation and react on changes based on configured thresholds and actions.

An *ActionTrigger* is made of the following attributes:

ID
	Identifier given by the administrator

UniqueID
	Per threshold identifier

ThresholdType
	Type of threshold configured. The following types are available:

	**\*min_balance**
		Matches when the :ref:`Balance` value is smaller.

	**\*max_balance**
		Matches when the :ref:`Balance` value is higher.

	**\*balance_expired**
		Matches if :ref:`Balance` is expired.

	**\*min_event_counter**
		Consider smaller aggregated values within event based on filters.

	**\*max_event_counter**
		Consider higher aggregated values within event based on filters.

	**\*min_balance_counter**
		Consider smaller :ref:`Balance` aggregated value based on filters.

	**\*max_balance_counter**
		Consider higher :ref:`Balance` aggregated value based on filters.

ThresholdValue
	The value of the threshold to match.

Recurrent
	Execute *ActionTrigger* multiple times.

MinSleep
	Sleep in between executes.

ExpirationDate
	Time when the *ActionTrigger* will expire.

ActivationDate
	Only consider the *ActionTrigger* starting with this time.

Balance
	Filters selecting the balance/-s to monitor.

Weight
	Priority in the chain. Higher values have more priority.

ActionsID
	:ref:`Action` profile to call on match.

MinQueuedItems
	Avoid false positives if the number of items hit is smaller than this.

Executed
	Marks the *ActionTrigger* as executed.

LastExecutionTime
	Time when the *ActionTrigger* was executed last.


.. _Action:

Action
------

Actions are routines executed on demand (ie. by one of the three subsystems: :ref:`SchedulerS`, :ref:`ThresholdS` or :ref:`ActionTriggers <ActionTrigger>`) or called by API by external scripts.

An \*Action has the following parameters:

ID
	*ActionSet* identifier.

ActionType
	The type of action to execute. Can be one of the following:

	**\*log**
		Creates an entry in the log (either syslog or stdout).

	**\*reset_triggers**
		Reset the matching :ref:`ActionTriggers <ActionTrigger>`

	**\*cdrlog**
		Creates a CDR entry (used for example when automatically charging DIDs). The content of the generated CDR entry can be customized within a special template which can be passed in *ExtraParameters* of the *Action*.

	**\*set_recurrent**
		Set the recurrent flag on the matching :ref:`ActionTriggers <ActionTrigger>`.

	**\*unset_recurrent**
		Unset the recurrent flag on the matching :ref:`ActionTriggers <ActionTrigger>`.

	**\*allow_negative**
		Set the *AllowNegative* flag on the :ref:`Balance`.

	**\*deny_negative**
		Unset the *AllowNegative* flag on the :ref:`Balance`.

	**\*reset_account**
		Re-init the :ref:`Account` by setting all of it's :ref:`Balance's Value <Balance>` to 0 and re-initialize counters and :ref:`ActionTriggers <ActionTrigger>`.

	**\*topup_reset**
		Reset the :ref:`Balance` matching the filters to 0 and add the top-up value to it.

	**\*topup**
		Add the value to the :ref:`Balance` matching the filters.

	**\*debit_reset**
		Reset the :ref:`Balance` matching the filters to 0 and debit the value from it.

	**\*debit**
		Debit the value from the :ref:`Balance` matching the filters.

	**\*transfer_balance**
		Transfers units between accounts' balances. It ensures both source and destination balances are of the same type and non-expired. Destination account and balance IDs, and optionally a reference value, are obtained from Action's ExtraParameters ``{"DestinationAccountID":"","DestinationBalanceID":""}``. If a reference value is specified, the transfer ensures the destination balance reaches this value. If the destination account is different from the source, it is locked during the transfer.

	**\*reset_counters**
		Reset the :ref:`Balance` counters (used by :ref:`ActionTriggers <ActionTrigger>`).

	**\*enable_account**
		Unset the :ref:`Account` *Disabled* flag.

	**\*disable_account**
		Set the :ref:`Account` *Disabled* flag.

	**\*http_post**
		Post data over HTTP protocol to configured HTTP URL.

	**\*http_post_async**
		Post data over HTTP protocol to configured HTTP URL without waiting for the feedback of the remote server.

	**\*mail_async**
		Send data to configured email address in extra parameters.

	**\*set_ddestinations**
		Update list of prefixes for destination ID starting with: *\*ddc* out of StatS. Used in scenarios like autodiscovery of homezone prefixes.

	**\*remove_account**
		Removes the matching account from the system.

	**\*remove_balance**
		Removes the matching :ref:`Balances <Balance>` out of the :ref:`Account`.

	**\*set_balance**
		Set the matching balances.

	**\*transfer_monetary_default**
		Transfer the value of the matching balances into the *\*default* one.

	**\*cgr_rpc**
		Call a CGRateS API over RPC connection. The API call will be defined as template within the *ExtraParameters*.
	
	**\*alter_sessions**
		Processes the *ExtraParameters* field from the action to construct a request for the ``SessionSv1.AlterSessions`` API call.
		The ExtraParameters field format is expected as follows:
		  - tenant
		  - filters: separated by "&".
		  - limit, specifying the maximum number of sessions to alter.
		  - APIOpts: set of key-value pairs (separated by "&").
		  - Event: set of key-value pairs (separated by "&").

	**\*force_disconnect_sessions**
		Processes the *ExtraParameters* field from the action to construct a request for the ``SessionSv1.ForceDisconnect`` API call.
		The ExtraParameters field format is expected as follows:
		  - tenant
		  - filters: separated by "&".
		  - limit, specifying the maximum number of sessions to disconnect.
		  - APIOpts: set of key-value pairs (separated by "&").
		  - Event: set of key-value pairs (separated by "&").

	**\*export**
		Will send the event that triggered the action to be processed by EEs

	**\*reset_threshold**
		Will reset the specified Threshold in the *ExtraParameters* field by writing inside it the ``Tenant:ID`` of the threshold.
	
	**\*reset_stat_queue**
		Will reset the specified StatQueue in the *ExtraParameters* field by writing inside it the ``Tenant:ID`` of the StatQueue.

	**\*topup_zero_negative**
		Set the the matching balances to topup value if they are negative.

	**\*set_expiry**
		Set the *ExpirationDate* for the matching balances.

	**\*publish_account**
		Publish the :ref:`Account` and each individual :ref:`Balance` to the :ref:`ThresholdS`.

	**\*publish_balance**
		Publish the matching :ref:`Balances <Balance>` to the :ref:`ThresholdS`.

	**\*remove_session_costs**
		Removes entries from the :ref:`StorDB.session_costs <StorDB>` table. Additional filters can be specified within the *ExtraParameters*.

	**\*remove_expired**
		Removes expired balances of type matching the filter.

	**\*reset_account_cdr**
		Creates the account out of last *CDR* saved in :ref:`StorDB` matching the account details in the filter. The *CDR* should contain *AccountSummary* within it's *CostDetails*.

	**\*remote_set_account**
		When an event triggers the action, the event will be used to set an account from a remote server using the URL provided in the *ExtraParameters* field.

	**\*dynamic_threshold** 
		Processes the *ExtraParameters* field from the action to construct a Threshold profile
		The ExtraParameters field format is expected as follows:
			0. Tenant 
			1. ID
			2. FilterIDs: separated by "&".
			3. ActivationInterval: separated by "&".
			4. MaxHits
			5. MinHits
			6. MinSleep
			7. Blocker
			8. Weight
			9. ActionIDs: separated by "&".
			10. Async
			11. EeIDs: separated by "&".
			12. APIOpts: set of key-value pairs (separated by "&").
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
   			<Tenant[0];Id[1];FilterIDs[2];ActivationInterval[3];MaxHits[4];MinHits[5];MinSleep[6];Blocker[7];Weight[8];ActionIDs[9];Async[10];EeIDs[11];APIOpts[12]>

	**\*dynamic_stats** 
		Processes the *ExtraParameters* field from the action to construct a StatQueueProfile
		The ExtraParameters field format is expected as follows:
			0. Tenant
			1. ID
			2. FilterIDs: separated by "&".
			3. ActivationInterval: separated by "&".
			4. QueueLength
			5. TTL
			6. MinItems
			7. Metrics: separated by "&".
			8. MetricFilterIDs: separated by "&".
			9. Stored
			10. Blocker
			11. Weight
			12. ThresholdIDs: separated by "&".
			13. APIOpts: set of key-value pairs (separated by "&").
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Tenant[0];Id[1];FilterIDs[2];ActivationInterval[3];QueueLength[4];TTL[5];MinItems[6];Metrics[7];MetricFilterIDs[8];Stored[9];Blocker[10];Weight[11];ThresholdIDs[12];APIOpts[13]>

	**\*dynamic_attribute** 
		Processes the *ExtraParameters* field from the action to construct a AttributeProfile
		The ExtraParameters field format is expected as follows:
			0. Tenant
			1. ID
			2. Context: separated by "&".
			3. FilterIDs: separated by "&".
			4. ActivationInterval: separated by "&".
		 	5. AttributeFilterIDs: separated by "&".
		 	6. Path
		 	7. Type
		 	8. Value: separated by "&".
			9. Blocker
			10. Weight
			11. APIOpts: set of key-value pairs (separated by "&").
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Tenant[0];ID[1];Contexts[2];FilterIDs[3];ActivationInterval[4];AttributeFilterIDs[5];Path[6];Type[7];Value[8];Blocker[9];Weight[10];APIOpts[11]>

	**\*dynamic_action_plan** 
		Processes the *ExtraParameters* field from the action to construct an ActionPlan
		The ExtraParameters field format is expected as follows:
			0. Id
			1. ActionsId
			2. TimingId
			3. Weight
			4. Overwrite
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Id[0];ActionsId[1];TimingId[2];Weight[3];Overwrite[4]>

	**\*dynamic_action_plan_accounts** 
		Processes the *ExtraParameters* field from the action to construct an ActionPlan with account ids
		The ExtraParameters field format is expected as follows:
			0. Id
			1. ActionsId
			2. TimingId
			3. Weight
			4. Overwrite
			5. Tenant:AccountIDs: separated by "&".
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Id[0];ActionsId[1];TimingId[2];Weight[3];Overwrite[4];Tenant:AccountIDs[5]>

	**\*dynamic_action** 
		Processes the *ExtraParameters* field from the action to construct a new Action
		The ExtraParameters field format is expected as follows:
			0. ActionsId
			1. Action
			2. ExtraParameters encapsulated by \f
			3. Filters: separated by "&".
			4. BalanceId
			5. BalanceType
			6. Categories: separated by "&".
			7. DestinationIds: separated by "&".
			8. RatingSubject
			9. SharedGroups: separated by "&".
		   	10. ExpiryTime
		   	11. TimingIds: separated by "&".
		   	12. Units
		   	13. BalanceWeight
		   	14. BalanceBlocker
		   	15. BalanceDisabled
		   	16. Weight
		   	17. Overwrite
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<ActionsId[0];Action[1];ExtraParameters[2];Filter[3];BalanceId[4];BalanceType[5];Categories[6];DestinationIds[7];RatingSubject[8];SharedGroup[9];ExpiryTime[10];TimingIds[11];Units[12];BalanceWeight[13];BalanceBlocker[14];BalanceDisabled[15];Weight[16]>

	**\*dynamic_destination** 
		Processes the *ExtraParameters* field from the action to construct a new Destination
		The ExtraParameters field format is expected as follows:
			0. Id
			1. Prefix: separated by "&".
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Id;Prefix>

	**\*dynamic_filter** 
		Processes the *ExtraParameters* field from the action to construct a Filter
		The ExtraParameters field format is expected as follows:
			0. Tenant
			1. ID
			2. Type
			3. Path
			4. Values: separated by "&".
			5. ActivationInterval: separated by "&".
		 	6. APIOpts: set of key-value pairs (separated by "&").
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Tenant[0];ID[1];Type[2];Path[3];Values[4];ActivationInterval[5];APIOpts[6]>

	**\*dynamic_route** 
		Processes the *ExtraParameters* field from the action to construct a RouteProfile
		The ExtraParameters field format is expected as follows:
			0. Tenant
			1. ID
			2. FilterIDs: separated by "&".
			3. ActivationInterval: separated by "&".
			4. Sorting
			5. SortingParameters: separated by "&".
			6. RouteID
			7. RouteFilterIDs: separated by "&".
			8. RouteAccountIDs: separated by "&".
			9. RouteRatingPlanIDs: separated by "&".
		   	10. RouteResourceIDs: separated by "&".
		   	11. RouteStatIDs: separated by "&".
		   	12. RouteWeight
		  	13. RouteBlocker
		   	14. RouteParameters
		   	15. Weight
		   	16. APIOpts: set of key-value pairs (separated by "&").
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Tenant[0];ID[1];FilterIDs[2];ActivationInterval[3];Sorting[4];SortingParameters[5];RouteID[6];RouteFilterIDs[7];RouteAccountIDs[8];RouteRatingPlanIDs[9];RouteResourceIDs[10];RouteStatIDs[11];RouteWeight[12];RouteBlocker[13];RouteParameters[14];Weight[15];APIOpts[16]>

	**\*dynamic_ranking** 
		Processes the *ExtraParameters* field from the action to construct a RankingProfile
		The ExtraParameters field format is expected as follows:
			0. Tenant
			1. ID
			2. Schedule
			3. StatIDs: separated by "&".
			4. MetricIDs: separated by "&".
			5. Sorting
			6. SortingParameters: separated by "&".
			7. Stored
			8. ThresholdIDs: separated by "&".
		    9. APIOpts: set of key-value pairs (separated by "&").
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Tenant[0];Id[1];Schedule[2];StatIDs[3];MetricIDs[4];Sorting[5];SortingParameters[6];Stored[7];ThresholdIDs[8];APIOpts[9]>

	**\*dynamic_rating_profile** 
		Processes the *ExtraParameters* field from the action to construct a RatingProfile
		The ExtraParameters field format is expected as follows:
			0. Tenant
			1. Category
			2. Subject
			3. ActivationTime
			4. RatingPlanId
			5. RatesFallbackSubject
		   	6. APIOpts: set of key-value pairs (separated by "&").
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Tenant[0];Category[1];Subject[2];ActivationTime[3];RatingPlanId[4];RatesFallbackSubject[5];APIOpts[6]>

	**\*dynamic_trend** 
		Processes the *ExtraParameters* field from the action to construct a TrendProfile
		The ExtraParameters field format is expected as follows:
			0. Tenant
			1. ID
			2. Schedule
			3. StatID
			4. Metrics: separated by "&".
			5. TTL
			6. QueueLength
			7. MinItems
			8. CorrelationType
			9. Tolerance
		   	10. Stored
		   	11. ThresholdIDs: separated by "&".
		   	12. APIOpts: set of key-value pairs (separated by "&").
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Tenant[0];Id[1];Schedule[2];StatID[3];Metrics[4];TTL[5];QueueLength[6];MinItems[7];CorrelationType[8];Tolerance[9];Stored[10];ThresholdIDs[11];APIOpts[12]>

	**\*dynamic_resource** 
		Processes the *ExtraParameters* field from the action to construct a ResourceProfile
		The ExtraParameters field format is expected as follows:
			0. Tenant
			1. Id
			2. FilterIDs: separated by "&".
			3. ActivationInterval: separated by "&".
			4. TTL
			5. Limit
			6. AllocationMessage
			7. Blocker
			8. Stored
			9. Weight
		   	10. ThresholdIDs: separated by "&".
		   	11. APIOpts: set of key-value pairs (separated by "&").
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Tenant[0];Id[1];FilterIDs[2];ActivationInterval[3];TTL[4];Limit[5];AllocationMessage[6];Blocker[7];Stored[8];Weight[9];ThresholdIDs[10];APIOpts[11]>

	**\*dynamic_action_trigger** 
		Processes the *ExtraParameters* field from the action to construct a ActionTrigger
		The ExtraParameters field format is expected as follows:
			0. Tag
			1. UniqueId
			2. ThresholdType
			3. ThresholdValue
			4. Recurrent
			5. MinSleep
			6. ExpiryTime
			7. ActivationTime
			8. BalanceTag
			9. BalanceType
		   	10. BalanceCategories: separated by "&".
		   	11. BalanceDestinationIds: separated by "&".
		   	12. BalanceRatingSubject
		   	13. BalanceSharedGroup: separated by "&".
		   	14. BalanceExpiryTime
		   	15. BalanceTimingIds: separated by "&".
		   	16. BalanceWeight
		   	17. BalanceBlocker
		   	18. BalanceDisabled
		   	19. ActionsId
		   	20. Weight
		Parameters are separated by ";" and must be provided in the specified order.

		.. code-block:: text
			
			<Tag[0];UniqueId[1];ThresholdType[2];ThresholdValue[3];Recurrent[4];MinSleep[5];ExpiryTime[6];ActivationTime[7];BalanceTag[8];BalanceType[9];BalanceCategories[10];BalanceDestinationIds[11];BalanceRatingSubject[12];BalanceSharedGroup[13];BalanceExpiryTime[14];BalanceTimingIds[15];BalanceWeight[16];BalanceBlocker[17];BalanceDisabled[18];ActionsId[19];Weight[20]>
		

Configuration
-------------

The *RALs* is configured within **rals** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

thresholds_conns
	Connections towards :ref:`ThresholdS` component, used for :ref:`Account` notifications.

stats_conns
	Connections towards :ref:`StatS` component, used for :ref:`Account` ralated metrics.

caches_conns
	Connections towards :ref:`CacheS` used for data reloads.

rp_subject_prefix_matching
	Enabling prefix matching for rating *Subject* field.

remove_expired
	Enable automatic removal of expired :ref:`Balances <Balance>`.

max_computed_usage
	Prevent usage rating calculations per type of records to avoid memory overload.

max_increments
	The maximum number of increments generated as part of rating calculations.

balance_rating_subject
	Default rating subject for balances, per balance type.


Use cases
---------

* Classic rater calculating costs for events using :ref:`Rating`.
* Account bundles for fixed and mobile networks (xG) using :ref:`Accounting`.
* Volume discounts in real-time using :ref:`Accounting`.
* Fraud detection with automatic mitigation using :ref:`ActionTriggers <ActionTrigger>`.