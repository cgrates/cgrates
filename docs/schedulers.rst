.. _SchedulerS:

SchedulerS
==========

**SchedulerS** is a standalone service component part of the **CGRateS** infrastructure, designed to execute account actions at scheduled times. It is configured via ActionPlans, with **FilterS** support to limit which accounts are targeted.

**SchedulerS** can be dynamically started and stopped via the ServiceManager.

Complete interaction with **SchedulerS** is possible via `CGRateS RPC APIs <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_.


Processing logic
----------------

An ActionPlan groups one or more ActionTimings, each specifying when a set of Actions should run and on which accounts. When **SchedulerS** starts or reloads, it fetches all ActionPlans from DataManager and builds a sorted priority queue of ActionTimings.

ActionTimings marked as *\*asap* are not queued and are executed when the ActionPlan is loaded. ActionTimings whose next start time is in the past are discarded.

The scheduler calculates the exact time until the next ActionTiming is due and sleeps until then. After execution, recurring ActionTimings are reinserted into the queue with their next calculated start time. ActionTimings with a *+* prefix in their StartTime (e.g. *+5m*) are executed once after the specified duration and then removed from the queue. A reload via API causes the loop to restart and rebuild the queue from the updated ActionPlans.

Filters defined in the configuration are applied during queue building, removing non-matching accounts from each ActionTiming before it is queued.


Parameters
----------

It is configured within the **schedulers** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

cdrs_conns
	Connections towards **CDRs** used for *\*cdrlog* actions. Possible values: <""\|*internal\|$rpc_conns_id>.

thresholds_conns
	Connections towards **ThresholdS** used for *\*reset_threshold* actions. Possible values: <""\|*internal\|$rpc_conns_id>.

stats_conns
	Connections towards **StatS** used for *\*reset_stat_queue* actions. Possible values: <""\|*internal\|$rpc_conns_id>.

filters
	List of filter IDs applied during queue building. Non-matching accounts are removed from each ActionTiming before it enters the queue.

dynaprepaid_actionplans
	List of ActionPlan IDs to be executed in case of *\*dynaprepaid* request type.


APIs logic
----------

SchedulerSv1.Reload
^^^^^^^^^^^^^^^^^^^

Triggers a full queue rebuild from the current ActionPlans in DataManager.


SchedulerSv1.ExecuteActions
^^^^^^^^^^^^^^^^^^^^^^^^^^^

Executes ActionPlan actions immediately, bypassing the schedule. Two modes are available and can be combined in a single call:

- **By ActionPlanID**: loads the ActionPlan and executes its ActionTimings sorted by weight. All accounts linked to the ActionPlan are affected.
- **By time interval**: collects all ActionPlans and executes their ActionTimings chronologically between TimeStart and TimeEnd.

ActionPlanID
	The ActionPlan to execute.

TimeStart
	Start of the time interval.

TimeEnd
	End of the time interval.


SchedulerSv1.ExecuteActionPlans
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Executes one or more ActionPlans for a single account. If the account does not exist, it is created.

ActionPlanIDs
	List of ActionPlan IDs to execute.

Tenant
	Tenant of the target account.

AccountID
	ID of the target account.


APIerSv1.GetScheduledActions
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

Returns the ActionTimings currently in the scheduler queue, with optional filtering and pagination. Tenant and Account filters apply only to ActionTimings that have accounts linked to them.

Tenant
	Return only ActionTimings with accounts matching this tenant.

Account
	Return only ActionTimings with accounts matching this account ID.

TimeStart
	Return only ActionTimings with NextRunTime at or after this value.

TimeEnd
	Return only ActionTimings with NextRunTime strictly before this value.

Limit
	Maximum number of results. All matching records are returned if not set.

Offset
	Number of records to skip before returning results.

**Each returned record contains:**

NextRunTime
	When this ActionTiming is next scheduled to run.

Accounts
	Number of accounts it will run on.

ActionPlanID
	The ActionPlan this ActionTiming belongs to.

ActionTimingUUID
	Unique identifier of the ActionTiming.

ActionsID
	The Actions set to be executed.


Use cases
---------

* Monthly balance top-ups or resets across a set of accounts.
* Trial account deactivation after a fixed delay using the *+* prefix.
* Replaying scheduled actions after a maintenance window using *TimeStart* / *TimeEnd*.
* Executing an ActionPlan immediately for a specific account via *ExecuteActionPlans*.
* Limiting scheduled action execution to specific accounts using :ref:`FilterS <FilterS>`.