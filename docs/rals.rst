.. _rals:

RALs
====


**RALs** is a standalone subsystem within **CGRateS** designed to handle two major tasks: :ref:`Rating` and :ref:`Accounting`. It is accessed via `CGRateS RPC APIs <https://godoc.org/github.com/cgrates/cgrates/apier/>`_.



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

Groups together rates per destination. Configured via the following parameters:

ID
	The tag uniquely idenfying each RatingPlan. There can be multiple entries grouped by the same ID.

DestinationRatesID
	The identifier of the :ref:`DestinationRate` set.

Weight
	Priority of matching rule (*DestinationRatesID*). Higher value equals higher priority.


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
	Matching all types of events after specific ones, represents generic units (ie: for each x *voice minutes, y *sms units, z *data units will have )

\*monetary
	Matching all types of events after specific ones, represents monetary units (can be interpreted as virtual currency).



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

Disabled
	Makes the *Balance* invisible to charging.

Factor
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

An *Action has the following parameters:

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

	**\*cdr_account**
		Creates the account out of last *CDR* saved in :ref:`StorDB` matching the account details in the filter. The *CDR* should contain *AccountSummary* within it's *CostDetails*.


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