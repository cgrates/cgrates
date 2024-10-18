.. _RankingS:


RankingS
========


**RankingS** is a standalone subsystem part of the **CGRateS** infrastructure, designed to work as an extension of the :ref:`StatS`, by regularly querying it for a list of predefined StatProfiles and ordering them based on their metrics.

Complete interaction with **RankingS** is possible via `CGRateS RPC APIs <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_.

Due it's real-time nature, **RankingS** are designed towards high throughput being able to process thousands of queries per second. This is doable since each *Ranking* is a very light object, held in memory and eventually backed up in :ref:`DataDB`.


Processing logic
----------------

In order for **RankingS** to start querying the :ref:`StatS`, it will need to be *scheduled* to do that. Scheduling is being done using `Cron Expressions <https://en.wikipedia.org/wiki/Cron>`_. 

Once *Cron Expressions* are defined within a *RankingProfile*, internal **Cron Scheduler** needs to be triggered. This can happen in two different ways:

Automatic Query Scheduling
	The profiles needing querying will be inserted into **RankingS** :ref:`JSON configuration <configuration>`. By leaving any part of *ranking_id* or *tenat* empty, it will be interpreted as catch-all filter.

API Scheduling
	The profiles needing querying will be sent inside arguments to the `RankingSv1.ScheduleQueries API call <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_.


Offline storage
---------------

Offline storage is optionally possible, by enabling profile *Stored* flag and configuring the *store_interval* inside :ref:`JSON configuration <configuration>`.


Ranking querying
----------------

In order to query a **Ranking** (ie: to be displayed in a web interface), one should use the `RankingSv1.GetRanking API call <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_ or `RankingSv1.GetRankingSummary API call <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`.


Ranking exporting
---------------

On each **Ranking** change, it will be possible to send a specially crafted *RankingSummary* event to one of the following subsystems:

**ThresholdS**
	Sending the **RankingUpdate** Event gives the administrator the possiblity to react to *Ranking* changes, including escalation strategies offered by the **TresholdS** paramters. 
	Fine tuning parameters (ie. selecting only specific ThresholdProfiles to increase speed) are available directly within the **TrendProfile**.

**EEs**
	**EEs** makes it possible to export the **RankingUpdate** to all the availabe outside interfaces of **CGRateS**.

Both exporting options are enabled within :ref:`JSON configuration <configuration>`.


Parameters
----------


RankingS
^^^^^^

**RankingS** is the **CGRateS** service component responsible of generating the **Ranking** queries. 

It is configured within **RankingS** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.

store_interval
	Time interval for backing up the RankingS into *DataDB*. 0 To completely disable the functionality, -1 to enable synchronous backup. Anything higher than 0 will give the interval for asynchronous backups.

stats_conns
	List of connections where we will query the stats.

scheduled_ids
	Limit the RankingProfiles generating queries towards **StatS**. Empty to enable all available RankingProfiles or just tenants for all the available profiles on a tenant.

thresholds_conns
	Connection IDs towards *ThresholdS* component. If not defined, there will be no notifications sent to *ThresholdS* on *Trend* changes.

ees_conns
	Connection IDs towards the *EEs* component. If left empty, no exports will be performed on *Trend* changes.

ees_exporter_ids
	Limit the exporter profiles executed on *Ranking* changes.


RankingProfile
^^^^^^^^^^^^^^

Ís made of the following fields:

Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	Identifier for the *RankingProfile*, unique within a *Tenant*.

Schedule
	Cron expression scheduling gathering of the metrics.

StatIDs
	List of **StatS** instances to query.

MetricIDs
	Limit the list of metrics from the stats instance queried.

Sorting
	Sorting strategy for the StatIDs. Possible values: 
	
	\*asc
		Sort the StatIDs ascendent based on list of MetricIDs provided in SortParameters. One or more MetricIDs can be specified in hte SortingParameters for the cases when one level sort is not enough to differentiate them. If all metrics will be equal, a random sort will be applied.	
	
	\*desc
		Sort the StatIDs descendat based on list of MetricIDs provided in SortParameters. One or more MetricIDs can be specified in hte SortingParameters for the cases when one level sort is not enough to differentiate them. If all metrics will be equal, a random sort will be applied.

SortingParameters 
	List of sorting parameters. For the current sorting strategies (\*asc/\*desc) there will be one or more MetricIDs defined. 
	Metric can be defined in compressed mode (ie. ["Metric1","Metric2"]) or extended mode (ie: ["Metric1:true", "Metric2:false"]) where *false* will reverse the sorting logic for that particular metric (ie: ["\*tcc:true","\*pdd:false"] with \*desc sorting strategy). 

Stored
	Enable storing of this *Ranking* intance for persistence.

ThresholdIDs
	Limit *TresholdProfiles* processing the *RankingUpdate* for this *RankingProfile*.


Ranking
^^^^^^^

instance is made out of the following fields:

Tenant 
	The tenant on the platform (one can see the tenant as partition ID).

ID 
	Unique *Ranking* identifier on a *Tenant*.

LastUpdate 
	Time of the last Metrics update.

Metrics
	Stat Metrics and their values at the query time.

Sorting
	Archived sorting strategy from the profile.

SortingParameters
	Archived list of sorted parameters from the profile.

SortedStatIDs
	List of queried stats, sorted based on sorting strategy and parameters.


Use cases
---------

* Ranking computation for commercial and monitoring applications.
* Revenue assurance applications.
* Fraud detection by ranking specific billing metrics during sensitive time intervals (\*acc, \*tcc, \*tcd).
* Building call patterns.
* Building statistical information to train systems capable of artificial intelligence.
* Building quality metrics used in traffic routing.


