.. _trends:

TrendS
=====


**TrendS** is a standalone subsystem part of the **CGRateS** infrastructure, designed to store *StatS* in a time-series-like database and calculate trend percentages based on their evolution.

Complete interaction with **TrendS** is possible via `CGRateS RPC APIs <https://pkg.go.dev/github.com/cgrates/cgrates/apier@master/>`_.

Due it's real-time nature, **StatS** are designed towards high throughput being able to process thousands of *Events* per second. This is doable since each *StatQueue* is a very light object, held in memory and eventually backed up in *DataDB*.


Processing logic
----------------




Parameters
----------


TrendS
^^^^^^

**TrendS** is the **CGRateS** component responsible of handling the *Trend* queries. 

It is configured within **trends** section from :ref:`JSON configuration <configuration>` via the following parameters:

enabled
	Will enable starting of the service. Possible values: <true|false>.



TrendProfile
^^^^^^^^^^^^

Ís made of the following fields:

Tenant
	The tenant on the platform (one can see the tenant as partition ID).

ID
	Identifier for the *TrendProfile*, unique within a *Tenant*.

FilterIDs
	List of *FilterProfileIDs* which should match in order to consider the profile matching the event.




Trend
^^^^^



Use cases
---------

* Aggregate various traffic metrics for traffic transparency.
* Revenue assurance applications.
* Fraud detection by aggregating specific billing metrics during sensitive time intervals (\*acc, \*tcc, \*tcd).
* Building call patterns.
* Building statistical information to train systems capable of artificial intelligence.
* Building quality metrics used in traffic routing.
