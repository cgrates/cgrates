1.Introduction
==============
CGRateS is a very fast and easy scalable rating engine targeted especially for ISPs and Telecom Operators.

It is written in Go (http://golang.org) and accessible from any language via JSON RPC. The code is well documented (go doc compliant API docs) and heavily tested.

After testing various databases like Kyoto_ cabinet, Redis_ or Mongodb_, the project focused on Redis as it delivers the best trade-off between speed, configuration and scalability. Despite that a connection to any database can be easily integrated by writing a simple adapter.

.. _kyoto: http://fallabs.com/kyotocabinet
.. _Redis: http://redis.io
.. _Mongodb: http://www.mongodb.org

To better understand the CGRateS architecture, bellow are some of the configurations in which CGRateS can operate:

.. image::  images/Simple.png

This scenario fits most of the simple installations. The Balancer can be left out and the Rater can be queried directly.

.. image::  images/Normal.png

While the network grows more Raters can be thrown into the stack to offer more requests per seconds workload. This implies the usage of the Balancer to distribute the requests to the Raters running on the different machines.

.. image::  images/Complicated.png

Of course more SessionManagers can serve multiple Telecom Switches and all of them are connected to the same Balancer. We are planning to support multiple Balancers for huge networks if the need arises.


1.1. CGRateS Features
---------------------
- Reliable and Fast ( very fast ;) ). To get an idea about speed, we have benchmarked 11000+ req/sec on a rather modest machine without requiring special tweaks in the kernel.
   - Using most modern programming concepts like multiprocessor support, asynchronous code execution within microthreads.
   - Built-in data caching system per call duration.
   - In-Memory database with persistence over restarts.
   - Use of Balancer assures High-Availability of Raters as well as increase of processing performance where that is required.
   - Use of Linux enterprise ready tools to assure High-Availability of the Balancer where that is required (*Supervise* for Application level availability and *LinuxHA* for Host level availability).
- Modular architecture
    - Easy to enhance functionality by rewriting custom session managers or mediators.
    - Flexible API accessible via both Gob (Golang specific, increased performance) or JSON (platform independent, universally accesible).
- Prepaid, Postpaid and Pseudo-Prepaid Controller.
    - Mutiple Primary Balances per Account (eg: MONETARY, SMS, INTERNET_MINUTES, INTERNET_TRAFFIC).
    - Multiple Auxiliary Balances per Account (eg: Free Minutes per Destination,  Volume Rates, Volume Discounts).
    - Concurrent sessions per account sharing the same balance with configurable debit interval (starting with 1 second).
    - Built-in Task-Scheduler supporting both one-time as well as recurrent actions (eg: TOPUP_MINUTES_PER_DESTINATION, DEBIT_MONETARY, RESET_BALANCE).
    - ActionTriggers ( useful for commercial offerings like receive amounts of monetary units if a specified number of minutes was charged in a month).
- Highly configurable Rating.
    - Connect Fees.
    - Priced Units definition.
    - Rate increments.
    - Millisecond timestaps.
    - Four decimal currencies.
    - Multiple TypeOfRecord rating (eg: standard vs. premium calls, SMSes, Internet Traffic).
    - Rating subject concatenations for combined records (eg: location based rating for same user).
    - Recurrent rates definition (per year, month, day, dayOfWeek, time).
    - Rating Profiles activation times (eg: rates becoming active at specific time in future).
- Multi-Tenant for both Prepaid as well as Rating.
- Flexible Mediator able to run multiple mediation processes on the same CDR.
- Verbose action logging in persistent databases (eg: Postgres) to cope with country specific law requirements.
- Good documentation ( that's me :).
- "Free as in Beer" with commercial support available on-demand.




