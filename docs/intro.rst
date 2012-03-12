Introduction
============
CGRates is a very fast and easy scalable rating engine targeted especially for telecom providers.

It is written in go (http://golang.net) and accessible from any language via JSON RPC. The code is well documented (go doc compliant API docs) and heavily tested.

Supported databases: kyoto_ cabinet, redis_, mongodb_.

.. _kyoto: http://fallabs.com/kyotocabinet
.. _redis: http://redis.io
.. _mongodb: http://www.mongodb.org

Features
--------
+ Rates for prepaid and for postpaid
+ The budget expressed in money and/or minutes (seconds)
+ High accuracy rating: configurable to milliseconds
+ Handles volume discount
+ Received calls bonus
+ Fully/Easy configurable 
+ Very fast (5000+ req/sec)
+ Good documentation
+ Commercial support available

How does CGRates work?
----------------------
Let's start with the most important function: finding the cost of a certain call. The call information comes to CGRates as the following values: subject, destination, start time and end time. The engine will lookup in the database for the activation periods applicable to the received subject and destination. What are the activation periods?

The activation period is a structure describing different prices for a call on different intervals of time. This structure has an activation time, from this time on the activation period is in effect and one ore more (usually more than one) intervals with prices. An interval is looking like this:

::

	type Interval struct {
		Month                                  time.Month
		MonthDay                               int
		WeekDays                               []time.Weekday
		StartTime, EndTime                     string // ##:##:## format
		Ponder, ConnectFee, Price, BillingUnit float64
	}

It specifies the Month, the MonthDay, the WeekDays and the StartTime and the EndTime when the Interval's Price is in effect. 

For example the interval {"Month": 1, "WeekDays":[1,2,3,4,5]. "StartTime":"18:00:00", "Price":0.1, "BillingUnit": 1} specifies that the Price for the first month of each year from Monday to Friday starting 18:00 is 0.1 cents per second. Most structure elements are optional and they can be combined in any way it makes sense. If an element is omitted it means it is zero ore any.

The ConnectFee specifies the connection price for the call if this interval is the first one from the call and the Ponder will establishes which interval will set the price for a call segment if more then one applies to it. 

The other functions relay on a user budget structure to manage the postpaid and prepaid different quotas.
