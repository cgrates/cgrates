4.2. Tariff Plans
=================
Major concept within CGRateS architecture, implement mechanisms to load rating as well as account data into CGRateS. 
For importing the data into CGRateS database(s) we are using **csv** *files*. 
The import process can be started as many times it is desired with one ore more csv files
and the existing values are overwritten.

.. important:: If **-flushdb** option is used when importing data with cgr-loader, 
               then the database **is cleaned** before importing. 

For more details see the **cgr-loader** tool from the tutorial chapter.

The rest of this section we will describe the content of every csv file.

4.2.1. Destinations
~~~~~~~~~~~~~~~~~~~
The destinations are binding together various prefixes / caller ids to define a
logical destination group. A prefix can appear in multiple destination groups.

::

    "Destinations.csv" - csv
    "tp_destinations"  - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/Destinations.csv
    :header-rows: 1

[0] - Id:
    Destination Id, a string by which this destination will be referenced in other places by.

[1] - Prefix:
    Prefix(es) attached to this destination.
    The prefix or caller id to be added to the specified destination.

4.2.2. Timings
~~~~~~~~~~~~~~
Holds time related definitions.
Describes the time periods that have different rates attached to them.

::

    "Timings.csv" - csv
    "tp_timings"  - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/Timings.csv
    :header-rows: 1

[0] - Tag:
    String by which this timing will be referenced in other places by.

[1] - Years:
    Integers separated by semicolons (;) specifying the years for this time period.
    
    **\*any** in case of always.

[2] - Months:
    Integers from 1=January to 12=December separated by semicolons (;) specifying the months for this time period.

    **\*any** in case of always (equivalent to 1;2;3;4;5;6;7;8;9;10;11;12).

[3] - MonthDays:
    Integers from 1 to 31 separated by semicolons (;) specifying the month days for this time period.

    **\*any** in case of always.

[4] - WeekDays:
    Integers from 1=Monday to 7=Sunday separated by semicolons (;) specifying the week days for this time period.

    **\*any** in case of always.

[5] - Time:
    The start time for this time period.
    
    If you set it to **\*asap** (was **\*now**) it will be replaced with the time of the data importing.

4.2.3. Rates
~~~~~~~~~~~~
Defines price groups for various destinations which will be associated to
various timings.

::

    "Rates.csv" - csv
    "tp_rates"  - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/Rates.csv
    :header-rows: 1


[0] - Id:
    Rate Id, a string by which this *rate* will be referenced in other places by.

[1] - ConnectFee:
    ConnectFee applied once the call is answered.
    The price to be charged once at the beginning of the call to the specified
    destination.

[2] - Rate:
    Number of billing units this rate applies to.
    The price for the billing unit expressed in cents.

[3] - RateUnit:
    The billing unit expressed in seconds.

[4] - RateIncrement:
    This rate will apply in increments of duration.
    The time gap for the rate

[5] - GroupIntervalStart:
    When the rate starts

.. seealso:: Rateincrement and GroupIntervalStart are when the calls has
   different rates in the timeframe. For example, the first 30 seconds of the
   calls has a rate of €0.1 and after that €0.2. The rate for this will the same
   TAG with two RateIncrements

4.2.4. Destination Rates
~~~~~~~~~~~~~~~~~~~~~~~
Attach rates to destinations. 

::

    "DestinationRates.csv" - csv
    "tp_destination_rates" - stor_db 

.. csv-table::
    :file: ../data/tariffplans/tutorial/DestinationRates.csv
    :header-rows: 1

[0] - Id:
    tbd

[1] - DestinationId:
    tbd

[2] - RatesTag:
    tbd

[3] - RoundingMethod:
    tbd

[4] - RoundingDecimals:
    tbd

[5] - MaxCost:
    tbd

[6] - MaxCostStrategy:
    tbd

4.2.5. Rating Plans
~~~~~~~~~~~~~~~~~~~

The *rating plan* makes the links between **Rating Profiles**, **Timings** and **Destination Rates** so each of them can be
described once and various combinations are made possible.

::

    "RatingPlans.csv" - csv
    "tp_rating_plans" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/RatingPlans.csv
    :header-rows: 1

[0] - Id:
    A string by which this *rating plan* will be referenced in other places by.

[1] - DestinationRatesId:
    The rating id/tag described in the **Destination rates** file. (*DestinationRates.csv* - **Id**)

[2] - TimingTag:
    The timing tag described in the **Timings** file. (*Timings.csv* - **Tag**)

[3] - Weight:
    If multiple timings cab be applied to a call the one with the lower weight
    wins. An example here can be the Christmas day: we can have a special timing
    for this day but the regular day of the week timing can also be applied to
    this day. The weight will differentiate between the two timings.


4.2.6. Rating profiles
~~~~~~~~~~~~~~~~~~~~~~
The *rating profile* **describes** the prices to be applied for various calls to
various destinations in various time frames. When a call is made the CGRateS
system will locate the rates to be applied to the call using the rating profiles.

::

    "RatingProfiles.csv" - csv
    "tp_rating_profiles" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/RatingProfiles.csv
    :header-rows: 1

[0] - Direction:
    Can be **\*in** or **\*out** for the INBOUND and OUTBOUND calls.

[1] - Tenant:
    Used to distinguish between carriers if more than one share the same database in the CGRates system.

[2] - Category:
    Type of record specifies the kind of transmission this rate profile applies to.

[3] - Subject:
    The client/user for who this profile is detailing the rates.

[4] - ActivationTime:
    Multiple rates timings/prices can be created for one profile with different
    activation times. When a call is made the appropriate profile(s) will be
    used to rate the call. So future prices can be defined here and the
    activation time can be set as appropriate.

[5] - RatingPlanId:
    The rating plan id/tag described in the **Rating Plans** file. (*RatingPlans.csv* - **Id**)

    This specifies the profile to be used in case the call destination.

[6] - RatesFallbackSubject:
    This specifies another profile to be used in case the call destination will
    not be found in the current profile. The same tenant, tor and direction will
    be used.

[7] - CdrStatQueueIds:
    The cdr stats id described in the **Cdr Stats** file. (*CdrStats.csv* - **Id**)

    Stat Queue associated with this account.


4.2.7. Account actions
~~~~~~~~~~~~~~~~~~~~~~

Describes the actions to be applied to the clients/users accounts. There are two
kinds of actions: timed and triggered. For the timed actions there is a
scheduler application that reads them from the database and executes them at the
appropriate timings. The triggered actions are executed when the specified
balance counters reach certain thresholds.

The accounts hold the various balances and counters to activate the triggered
actions for each the client.

Balance types are: MONETARY, SMS, INTERNET, INTERNET_TIME, MINUTES.

::

    "AccountActions.csv" - csv
    "tp_account_actions" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/AccountActions.csv
    :header-rows: 1

[0] - Tenant:
    Used to distinguish between carriers if more than one share the same
    database in the CGRates system.

[1] - Account:
    The identifier for the user's account.

[2] - Direction:
    Can be **\*in** or **\*out** for the INBOUND and OUTBOUND calls.

[3] - ActionPlanId:
    The action plan id/tag described in the **Action plans** file. (*ActionPlans.csv* - **Id**)

    Forwards to a timed action group that will be used on this account.

[4] - ActionTriggersId:
    The action trigger id/tag described in the **Action triggers** file. (*ActionTriggers.csv* - **Tag**)

    Forwards to a triggered action group that will be applied to this account.

[5] - AllowNegative:
    TBD

[6] - Disabled:
    TBD

4.2.8 Action triggers
~~~~~~~~~~~~~~~~~~~~~~
For each account there are counters that record the activity on various
balances. Action triggers allow when a counter reaches a threshold to activate a
group of actions. After the execution the action trigger is marked as used and
will no longer be evaluated until the triggers are reset. See actions for action
trigger resetting.

::

    "ActionTriggers.csv" - csv
    "tp_action_triggers" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/ActionTriggers.csv
    :header-rows: 1

[0] - Tag:
    A string by which this action trigger will be referenced in other places by.

[1] - UniqueID:
    Unique id for the trigger in multiple ActionTriggers

[2] - ThresholdType:
    The threshold type. Can have one of the following:

    + **\*min_counter**: Fire when counter is less than ThresholdValue
    + **\*max_counter**: Fire when counter is greater than ThresholdValue
    + **\*min_balance**: Fire when balance is less than ThresholdValue
    + **\*max_balance**: Fire when balances is greater than ThresholdValue
    + **\*min_asr**: Fire when ASR(Average success Ratio) is less than ThresholdValue
    + **\*max_asr**: Fire when ASR is greater than ThresholdValue
    + **\*min_acd**: Fire when ACD(Average call Duration) is less than ThresholdValue
    + **\*max_acd**: Fire when ACD is greater than ThresholdValue
    + **\*min_acc**: Fire when ACC(Average call cost) is less than ThresholdValue
    + **\*max_acc**: Fire when ACC is greater than ThresholdValue
    + **\*min_tcc**: Fire when TCC(Total call cost) is less than ThresholdValue
    + **\*max_tcc**: Fire when TCC is greater than ThresholdValue
    + **\*min_tcd**: fire when TCD(total call duration) is less than thresholdvalue
    + **\*max_tcd**: fire when TCD is greater than thresholdvalue
    + **\*min_pdd**: Fire when PDD(Post Dial Delay) is less than ThresholdValue
    + **\*max_pdd**: Fire when PDD is greater than ThresholdValue

[3] - ThresholdValue:
    The value of the balance counter that will trigger this action.

[4] - Recurrent(Boolean):
    In case of trigger we can fire recurrent while it's active, or only the
    first time.

[5] - MinSleep:
    When Threshold is triggered we can sleep for the time specified.

[6] - ExpiryTime
    TBD

[7] - ActivationTime
    TBD

[8] - BalanceTag:
    Specifies the balance counter by which this action will be triggered. 
    Can be:

    + **MONETARY**
    + **SMS**
    + **INTERNET**
    + **INTERNET_TIME**
    + **MINUTES**

[9] - BalanceType:
    Specifies the balance type for this action:

    + **\*voice**:  units of call minutes
    + **\*sms**: units of SMS
    + **\*data**: units of data
    + **\*monetary**: units of money

[10] - BalanceDirections:
    Can be **\*in** or **\*out** for the INBOUND and OUTBOUND calls.

[11] - BalanceCategories:
    Category of the call/trigger

[12] - BalanceDestinationIds:
    The destination id/tag described in the **Destinations** file. (*Destinations.csv* - **Id**) - rinor: need verification

    Destination of the call/trigger

[13] - BalanceRatingSubject:
    TBD

[14] - BalanceSharedGroup:
    Shared Group of the call/trigger

[15] - BalanceExpiryTime:
    TBD

[16] - BalanceTimingIds:
    TBD

[17] - BalanceWeight:
    TBD

[18] - BalanceBlocker
    TBD

[19] - BalanceDisabled:
    TBD

[20] - StatsMinQueuedItems:
    Min of items that need to have a queue to reach this Trigger.
    Trigger actions only if this number is hit (stats only).

[21] - ActionsId:
    The actions id/tag described in the **Actions** file. (*Actions.csv* - **ActionsId**)

    Forwards to an action group to be executed when the threshold is reached.

[22] - Weight:
    Specifies the order for these triggers to be evaluated. If there are
    multiple triggers are fired in the same time the ones with the lower weight
    will be executed first.

4.2.9. Action Plans
~~~~~~~~~~~~~~~~~~~
TBD

::

    "ActionPlans.csv"  - csv
    "tp_account_plans" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/ActionPlans.csv
    :header-rows: 1

[0] - Id:
    A string by which this action timing will be referenced in other places by.

[1] - ActionsId:
    Forwards to an action group to be executed when the timing is right.

[2] - TimingId:
    A timing (one time or recurrent) at which the action group will be executed

[3] - Weight:
    Specifies the order for these timings to be evaluated. If there are multiple
    action timings set to be execute on the same time the ones with the lower
    weight will be executed first.

4.2.10. Actions
~~~~~~~~~~~~~~
TBD

::

    "Actions.csv" - csv
    "tp_actions"  - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/Actions.csv
    :header-rows: 1

[0] - ActionsId:
    A string by which this action will be referenced in other places by.

[1] - Action:
    The action type. Can have one of the following:

    + **\*allow_negative**: Allow to the account to have negative balance
    + **\*call_url**: Send a http request to the following url
    + **\*call_url_async**: Send a http request to the following url Asynchronous
    + **\*cdrlog**: Log the current action in the storeDB
    + **\*debit**: Debit account balance.
    + **\*deny_negative**: Deny to the account to have negative balance
    + **\*disable_account**: Disable account in the platform
    + **\*enable_account**: Enable account in the platform
    + **\*log**: Logs the other action values (for debugging purposes).
    + **\*mail_async**: Send a email to the direction
    + **\*reset_account**: Sets all counters to 0
    + **\*reset_counter**: Sets the counter for the BalanceTag to 0
    + **\*reset_counters**: Sets *all* the counters for the BalanceTag to 0
    + **\*reset_triggers**: reset all the triggers for this account
    + **\*set_recurrent**: (pending)
    + **\*topup**: Add account balance. If the specific balance is not defined, define it (example: minutes per destination).
    + **\*topup_reset**:  Add account balance. If previous balance found of the same type, reset it before adding.
    + **\*unset_recurrent**: (pending)
    + **\*unlimited**: (pending)

[2] - ExtraParameters:
    In Extra Parameter field you can define an argument for the action. In case
    of call_url Action, extraParameter will be the url action. In case of
    mail_async the email that you want to receive.

[3] - Filter
    TBD

[4] - BalanceId:
    The balance on which the action will operate

[5] - BalanceType:
    Specifies the balance type for this action:

    + **\*voice**:  units of call minutes
    + **\*sms**: units of SMS
    + **\*data**: units of data
    + **\*monetary**: units of money

[6] - Directions:
    Can be **\*in** or **\*out** for the INBOUND and OUTBOUND calls.

[7] - Categories:
    TBD

[8] - DestinationIds:
    The destination id/tag described in the **Destinations** file. (*Destinations.csv* - Id) 

    This field is used only if the BalanceId is MINUTES. Specifies the
    destination of the minutes to be operated.

[9] - RatingSubject:
    The ratingSubject of the Actions

[10] - SharedGroup:
    In case of the account uses any shared group for the balances.

[11] - ExpiryTime:
    TBD

[12] - TimingIds:
    Timming tag when the action can be executed. Default ALL.

[13] - Units:
    Number of units for decrease the balance. Only use if BalanceType is voice.

[14] - BalanceWeight:
    TBD

[15] - BalanceBlocker
    TBD

[16] - BalanceDisabled:
    TBD

[17] - Weight:
    If there are multiple actions in a group, they will be executed in the order
    of their weight (**smaller** first).

4.2.11. Derived Chargers
~~~~~~~~~~~~~~~~~~~~~~~~~
For each call we can bill more than one time, for that we need to use the
following options:

::

    "DerivedChargers.csv" - csv
    "tp_derived_chargers" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/DerivedChargers.csv
    :header-rows: 1

In derived charges we have 2 different kind of options, **FILTERS** and **ACTIONS** :

**Filters**: With the following fields we filter the calls that need to run a extra
billing parameter.

[0] - Direction:
    TBD
[1] - Tenant:
    TBD
[2] - Category:
    TBD
[3] - Account:
    TBD
[4] - Subject:
    TBD
[5] - DestinationIds:
    TBD

**Actions**: In case of the filter options match, platform creates extra runid with
the fields that we want to modify.

[6] - RunId:
    TBD
[7] - RunFilter:
    TBD
[8] - ReqTypeField:
    TBD
[9] - DirectionField:
    TBD
[10] - TenantField:
    TBD
[11] - CategoryField:
    TBD
[12] - AccountField:
    TBD
[13] - SubjectField:
    TBD
[14] - DestinationField:
    TBD
[15] - SetupTimeField:
    TBD
[16] - PddField:
    TBD
[17] - AnswerTimeField:
    TBD
[18] - UsageField:
    TBD
[19] - SupplierField:
    TBD
[20] - DisconnectCause:
    TBD
[21] - RatedField:
    TBD
[22] - CostField:
    TBD

In the example, all the calls with direction=out, tenant=cgrates.org,
category="call" and account and subject equal 1001. Will be created a new cdr in
the table *rated_cdrs* with the runID derived_run1, and the subject 1002.

This feature it's useful in the case that you want to rated the calls 2 times,
for example rated for different tenants or resellers.

4.2.12. CDR Stats
~~~~~~~~~~~~~~~~~~
CDR Stats enables some realtime statistics in your platform for multiple
purposes, you can read more, see :ref:`cdrstats-main`

::

    "CdrStats.csv" - csv
    "tp_cdr_stats" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/CdrStats.csv
    :header-rows: 1

[0] - Id:
    Tag name for the Queue id

[1] - QueueLength:
    Maximum number of calls in this queue

[2] - TimeWindow:
    Window frame to store the calls

[3] - SaveInterval:
    Each interval queue stats will save in the stordb

[4] - Metric:
    Type of metric see :ref:`cdrstats-metrics`

[5] - SetupInterval:
    TBD

[6] - TOR:
    TBD

[7] - CdrHost
    TBD

[8] - CdrSource:
    TBD

[9] - ReqType:
    Filter by reqtype

[10] - Direction:
    TBD

[11] - Tenant:
    Used to distinguish between carriers if more than one share the same
    database in the CGRates system.

[12] - Category:
    Type of record specifies the kind of transmission this rate profile applies
    to.

[13] - Account:
    The identifier for the user's account.

[14] - Subject:
    The client/user for who this profile is detailing the rates.

[15] - DestinationIds:
    Filter only by destinations prefix. Can be multiple separated with **;**

[16] - PddInterval:
    TBD

[17] - UsageInterval:
    TBD

[18] - Supplier:
    TBD

[19] - DisconnectCause:
    TBD

[20] - RunIds:
    TBD

[21] - RatedAccount:
    Filter by rated account

[22] - RatedSubject:
    Filter by rated subject

[23] - CostInterval:
    Filter by cost

[24] - ActionTriggers:
    ActionTriggers associated with this queue

4.2.13. Shared groups
~~~~~~~~~~~~~~~~~~~~~
TBD

::

    "SharedGroups.csv" - csv
    "tp_shared_groups" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/SharedGroups.csv
    :header-rows: 1

[0] - Id:
    TBD

[1] - Account:
    TBD

[2] - Strategy:
    TBD

[3] - RatingSubject:
    TBD

4.2.14. LCR rules
~~~~~~~~~~~~~~~~~
TBD

::

    "LcrRules.csv" - csv
    "tp_lcr_rules" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/LcrRules.csv
    :header-rows: 1

[0] - Direction:
    TBD

[1] - Tenant:
    TBD

[2] - Category:
    TBD

[3] - Account:
    TBD

[4] - Subject:
    TBD

[5] - DestinationTag:
    TBD

[6] - RpCategory:
    TBD

[7] - Strategy:
    TBD

[8] - StrategyParams:
    TBD

[9] - ActivationTime:
    TBD

[10] - Weight:
    TBD

4.2.15. Users
~~~~~~~~~~~~~
TBD

::

    "Users.csv" - csv
    "tp_users"  - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/Users.csv
    :header-rows: 1

[0] - Tenant:
   TBD

[1] - UserName:
   TBD

[2] - Masked:
   TBD

[3] - AttributeName:
   TBD

[4] - AttributeValue:
   TBD

[5] - Weight:
   TBD

4.2.16. Aliases
~~~~~~~~~~~~~~~
TBD

::

    "Aliases.csv" - csv
    "tp_aliases"  - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/Aliases.csv
    :header-rows: 1

[0] - Direction:
   TBD

[1] - Tenant:
   TBD

[2] - Category:
   TBD

[3] - Account:
   TBD

[4] - Subject:
   TBD

[5] - DestinationId:
   TBD

[6] - Context:
   TBD

[7] - Target:
   TBD

[8] - Original:
   TBD

[9] - Alias:
   TBD

[10] - Weight:
   TBD

4.2.17. Resource Limits
~~~~~~~~~~~~~~~~~~~~~~~
TBD

::

    "ResourceLimits.csv" - csv
    "tp_resource_limits" - stor_db

.. csv-table::
    :file: ../data/tariffplans/tutorial/ResourceLimits.csv
    :header-rows: 1

[0] - Tag
   TBD

[1] - FilterType
   TBD

[2] - FilterFieldName
   TBD

[3] - FilterValues
   TBD

[4] - ActivationTime
   TBD

[5] - Weight
   TBD

[6] - Limit
   TBD

[7] - ActionTriggerIds
   TBD

