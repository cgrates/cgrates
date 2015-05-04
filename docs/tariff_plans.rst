Tariff Plans
============

For importing the data into CGRateS database we are using cvs files. The import process can be started as many times it is desired with one ore more csv files and the existing values are overwritten. If the -flush option is used then the database is cleaned before importing.For more details see the cgr-loader tool from the tutorial chapter.

The rest of this section we will describe the content of every csv files.

4.2.1. Rates profile
~~~~~~~~~~~~~~~~~~~~

The rates profile describes the prices to be applied for various calls to various destinations in various time frames. When a call is made the CGRateS system will locate the rates to be applied to the call using the rating profiles.

.. csv-table::
    :file: ../data/tariffplans/tutorial/RatingProfiles.csv
    :header-rows: 1

Tenant
    Used to distinguish between carriers if more than one share the same database in the CGRates system.
TOR
    Type of record specifies the kind of transmission this rate profile applies to.
Direction
    Can be IN or OUT for the INBOUND and OUTBOUND calls.
Subject
    The client/user for who this profile is detailing the rates.
RatesFallbackSubject
    This specifies another profile to be used in case the call destination will not be found in the current profile. The same tenant, tor and direction will be used.
RatesTimingTag
    Forwards to a tag described in the rates timing file to be used for this profile.
ActivationTime
    Multiple rates timings/prices can be created for one profile with different activation times. When a call is made the appropriate profile(s) will be used to rate the call. So future prices can be defined here and the activation time can be set as appropriate.

4.2.2. Rating Plans
~~~~~~~~~~~~~~~~~~~

This file makes links between a ratings and timings so each of them can be described once and various combinations are made possible.

.. csv-table::
    :file: ../data/tariffplans/tutorial/RatingPlans.csv
    :header-rows: 1

Tag
    A string by which this rates timing will be referenced in other places by.
RatesTag
    The rating tag described in the rates file.
TimingTag
    The timing tag described in the timing file
Weight
    If multiple timings cab be applied to a call the one with the lower weight wins. An example here can be the Christmas day: we can have a special timing for this day but the regular day of the week timing can also be applied to this day. The weight will differentiate between the two timings.


4.2.3. Rates
~~~~~~~~~~~~
Defines price groups for various destinations which will be associated to various timings.


.. csv-table::
    :file: ../data/tariffplans/tutorial/Rates.csv
    :header-rows: 1


Tag
    A string by which this rate will be referenced in other places by.
DestinationsTag
    The destination tag which these rates apply to.
ConnectFee
    The price to be charged once at the beginning of the call to the specified destination.
Price
    The price for the billing unit expressed in cents.
BillingUnit
    The billing unit expressed in seconds

4.2.4. Timings
~~~~~~~~~~~~~~
Describes the time periods that have different rates attached to them.

.. csv-table::
    :file: ../data/tariffplans/tutorial/Timings.csv
    :header-rows: 1

Tag
    A string by which this timing will be referenced in other places by.
Months
    Integers from 1=January to 12=December separated by semicolons (;) specifying the months for this time period.
MonthDays
    Integers from 1 to 31 separated by semicolons (;) specifying the month days for this time period.
WeekDays
    Integers from 1=Monday to 7=Sunday separated by semicolons (;) specifying the week days for this time period.
StartTime
    The start time for this time period. \*now will be replaced with the time of the data importing.

4.2.5. Destinations
~~~~~~~~~~~~~~~~~~~

The destinations are binding together various prefixes / caller ids to define a logical destination group. A prefix can appear in multiple destination groups.

.. csv-table::
    :file: ../data/tariffplans/tutorial/Destinations.csv
    :header-rows: 1
Tag
    A string by which this destination will be referenced in other places by.
Prefix
    The prefix or caller id to be added to the specified destination.

4.2.6. Account actions
~~~~~~~~~~~~~~~~~~~~~~

Describes the actions to be applied to the clients/users accounts. There are two kinds of actions: timed and triggered. For the timed actions there is a scheduler application that reads them from the database and executes them at the appropriate timings. The triggered actions are executed when the specified balance counters reach certain thresholds.

The accounts hold the various balances and counters to activate the triggered actions for each the client.

Balance types are: MONETARY, SMS, INTERNET, INTERNET_TIME, MINUTES.

.. csv-table::
    :file: ../data/tariffplans/tutorial/AccountActions.csv
    :header-rows: 1

Tenant
    Used to distinguish between carriers if more than one share the same database in the CGRates system.
Account
    The identifier for the user's account.
Direction
    Can be IN or OUT for the INBOUND and OUTBOUND calls.
ActionTimingsTag
    Forwards to a timed action group that will be used on this account.
ActionTriggersTag
    Forwards to a triggered action group that will be applied to this account.

4.2.7 Action triggers
~~~~~~~~~~~~~~~~~~~~~~

For each account there are counters that record the activity on various balances. Action triggers allow when a counter reaches a threshold to activate a group of actions. After the execution the action trigger is marked as used and will no longer be evaluated until the triggers are reset. See actions for action trigger resetting.

.. csv-table::
    :file: ../data/tariffplans/tutorial/ActionTriggers.csv
    :header-rows: 1

Tag
    A string by which this action trigger will be referenced in other places by.

ThresholdType
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

ThresholdValue
    The value of the balance counter that will trigger this action.

Recurrent(Boolean)
    In case of trigger we can fire recurrent while it's active, or only the first time.

MinSleep
    When Threshold is triggered we can sleep for the time specified.

BalanceTag
    Specifies the balance counter by which this action will be triggered. Can be:

    + **MONETARY**
    + **SMS**
    + **INTERNET**
    + **INTERNET_TIME**
    + **MINUTES**

BalanceType
    Specifies the balance type for this action:

    + **\*voice**:  units of call minutes
    + **\*sms**: units of SMS
    + **\*data**: units of data
    + **\*monetary**: units of money

BalanceDirection
    Can be **\*in** or **\*out** for the INBOUND and OUTBOUND calls.

BalanceCategory

BalanceDestinationTag

BalanceRatingSubject

BalanceSharedGroup

BalanceExpiryTime

BalanceTimingTags

BalanceWeight

StatsMinQueuedItems

ActionsTag
    Forwards to an action group to be executed when the threshold is reached.

Weight
    Specifies the order for these triggers to be evaluated. If there are multiple triggers are fired in the same time the ones with the lower weight will be executed first.


DestinationTag
    This field is used only if the balanceTag is MINUTES. If the balance counter monitors call minutes this field indicates the destination of the calls for which the minutes are recorded.a

4.2.8. Action Plans
~~~~~~~~~~~~~~~~~~~

.. csv-table::
    :file: ../data/tariffplans/tutorial/ActionPlans.csv
    :header-rows: 1

Tag
    A string by which this action timing will be referenced in other places by.
ActionsTag
    Forwards to an action group to be executed when the timing is right.
TimingTag
    A timing (one time or recurrent) at which the action group will be executed
Weight
    Specifies the order for these timings to be evaluated. If there are multiple action timings set to be execute on the same time the ones with the lower weight will be executed first.

4.2.9. Actions
~~~~~~~~~~~~~~


.. csv-table::
    :file: ../data/tariffplans/tutorial/Actions.csv
    :header-rows: 1


Tag
    A string by which this action will be referenced in other places by.
Action
    The action type. Can have one of the following:

    + LOG: Logs the other action values (for debugging purposes).
    + RESET_TRIGGERS: Marks all action triggers as ready to be executed.
    + SET_POSTPAID: Sets account to postpaid, maintains it's balances.
    + RESET_POSTPAID: Set account to postpaid, reset all it's balances.
    + SET_PREPAID: Sets account to prepaid, maintains it's balances. Makes sense after an account was set to POSTPAID and admin wants it back.
    + RESET_PREPAID: Set account to prepaid, reset all it's balances.
    + TOPUP_RESET:  Add account balance. If previous balance found of the same type, reset it before adding.
    + TOPUP: Add account balance. If the specific balance is not defined, define it (eg: minutes per destination).
    + DEBIT: Debit account balance.
    + RESET_COUNTER: Sets the counter for the BalanceTag to 0
    + RESET_ALL_COUNTERS: Sets all counters to 0

BalanceTag
    The balance on which the action will operate
Units
    The units which will be operated on the balance BalanceTag.
DestinationTag
    This field is used only if the balanceTag is MINUTES. Specifies the destination of the minutes to be operated.
PriceType
    This field is used only if the balanceTag is MINUTES. Specifies if the minutes price will be absolute or a percent of the normal price, Can be ABSOLUTE or PERCENT. If the value is percent the
PriceValue
    This field is used only if the balanceTag is MINUTES. The price for each second.
MinutesWeight
    This field is used only if the balanceTag is MINUTES. If more minute balances are suitable for a call the one with smaller weight will be used first.
Weight
    If there are multiple actions in a group, they will be executed in the order of their weight (smaller first).


4.2.10. Derived Chargers
~~~~~~~~~~~~~~~~~~~~~~~~~

For each call we can bill more than one time, for that we need to use the
following options:

.. csv-table::
    :file: ../data/tariffplans/tutorial/DerivedChargers.csv
    :header-rows: 1

In derived charges we have 2 different kind of options, filters, and actions:

Filters: With the following fields we filter the calls that need to run a extra
billing parameter.
    + Direction
    + Tenant
    + Category
    + Account
    + Subject

Actions: In case of the filter options match, platform creates extra runid with
the fields that we want to modify.

    + RunId
    + RunFilter
    + ReqTypeField
    + DirectionField
    + TenantField
    + CategoryField
    + AccountField
    + SubjectField
    + DestinationField
    + SetupTimeField
    + AnswerTimeField
    + UsageField

In the example, all the calls with direction=out, tenant=cgrates.org,
category="call" and account and subject equal 1001. Will be created a new cdr in
the table *rated_cdrs* with the runID derived_run1, and the subject 1002.

This feature it's useful in the case that you want to rated the calls 2 times,
for example rated for different tenants or resellers.
