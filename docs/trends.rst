.. _trends:

TrendS
=====


**TrendS** is a standalone subsystem part of the **CGRateS** infrastructure, designed to work as an extension of the :ref:`StatS`, by regularly querying it, storing it's values in a time-series-like database and calculate trend percentages based on their evolution.

Complete interaction with **TrendS** is possible via `CGRateS RPC APIs <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_.

Due it's real-time nature, **TrendS** are designed towards high throughput being able to process thousands of queries per second. This is doable since each *Trend* is a very light object, held in memory and eventually backed up in :ref:`DataDB`.


Processing logic
----------------

In order for **TrendS** to start querying the :ref:`StatS`, it will need to be *scheduled* to do that. Scheduling is being done using `Cron Expressions <https://en.wikipedia.org/wiki/Cron>`_. 

Once *Cron Expressions* are defined within a *TrendProfile*, internal **Cron Scheduler** needs to be triggered. This can happen in two different ways:

Automatic Query Scheduling
	The profiles needing querying will be inserted into **trends** :ref:`JSON configuration <configuration>`. By leaving any part of *trend_id* or *tenat* empty, it will be interpreted as catch-all filter.

API Scheduling
	The profiles needing querying will be sent inside arguments to the `TrendSv1.ScheduleQueries API call <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_.


Offline storage
---------------

Offline storage is optionally possible, by enabling profile *Stored* flag and configuring the *store_interval* inside :ref:`JSON configuration <configuration>`.


Trend querying
--------------

In order to query a **Trend** (ie: to be displayed in a web interface), one should use the `TrendSv1.GetTrend API call <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_ which also offers pagination parameters.


Trend exporting
---------------

On each **Trend** change, it will be possible to send a specially crafted event, *TrendUpdate* to one of the following subsystems:

**ThresholdS**
	Sending the **TrendUpdate** Event gives the administrator the possiblity to react to *Trend* changes, including escalation strategies offered by the **TresholdS** paramters. 
	Fine tuning parameters (ie. selecting only specific ThresholdProfiles to increase speed1) are available directly within the **TrendProfile**.

**EEs**
	**EEs** makes it possible to export the **TrendUpdate** to all the availabe outside interfaces of **CGRateS**.

Both exporting options are enabled within :ref:`JSON configuration <configuration>`.


Parameters
----------


TrendS
^^^^^^

**TrendS** is the **CGRateS** component responsible of generating the **Trend** queries. 

It is configured within **trends** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

store_interval
	Time interval for backing up the trends into *DataDB*. 0 To completely disable the functionality, -1 to enable synchronous backup. Anything higher than 0 will give the interval for asynchronous backups.

stats_conns
	List of connections where we will query the stats.

scheduled_ids
	Limit the TrendProfiles queried. Empty to query all the available TrendProfiles or just tenants for all the available profiles on a tenant.

thresholds_conns
	Connection IDs towards *ThresholdS* component. If not defined, there will be no notifications sent to *ThresholdS* on *Trend* changes.

ees_conns
	Connection IDs towards the *EEs* component. If left empty, no exports will be performed on *Trend* changes.

ees_exporter_ids
	Limit the exporter profiles executed on *Trend* changes.


TrendProfile
^^^^^^^^^^^^

Ís made of the following fields:

Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	Identifier for the *TrendProfile*, unique within a *Tenant*.

Schedule
	Cron expression scheduling gathering of the metrics.

StatID
	StatS identifier which will be queried.

Metrics
	Limit the list of metrics from the stats instance queried.

TTL 
	Automatic cleanup of the queried values from inside *Trend Metrics*.
	
QueueLength 
	Limit the size of *Trend Metrics*. Older values will be removed first.

MinItems 
	Issue *TrendUpdate* events to external subsystems only if MinItems are reched to limit false alarms.

CorrelationType 
	The correlation strategy to use when computing the trend. *\*average* will consider all previous query values and *\*last* only the last one.

Tolerance
	Allow a deviation of the values when computin the trend. This is defined as percentage of increase/decrease.

Stored
	Enable storing of this *Trend* for persistence.

ThresholdIDs
	Limit *TresholdProfiles* processing the *TrendUpdate* for this *TrendProfile*.


Trend
^^^^^

is made out of the following fields:

Tenant 
	The tenant on the platform (one can see the tenant as partition ID).

ID 
	Unique *Trend* identifier on a *Tenant*

RunTimes 
	Times when the stat queries were ran by the scheduler

Metrics
	History of the queried metrics, indexed by the query time. One query stores the following values:

	ID 
		Metric ID on the *StatS* side 

	Value 
		Value of the metric at the time of query 

	TrendGrowth 
		Computed trend growth for the metric values, stored in percentage numbers.

	TrendLabel 
		Computed trend label for the metric values. Possible values are: *positive, *negative, *constant, N/A.


Use cases
---------

* Aggregate various traffic metrics for traffic transparency.
* Revenue assurance applications.
* Fraud detection by aggregating specific billing metrics during sensitive time intervals (\*acc, \*tcc, \*tcd).
* Building call patterns.
* Building statistical information to train systems capable of artificial intelligence.
* Building quality metrics used in traffic routing.
