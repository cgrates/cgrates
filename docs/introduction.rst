1. Introduction
===============

`CGRateS`_ is a *very fast* and *easily scalable* **(charging, rating, accounting, lcr, mediation, billing, authorization)** *ENGINE* targeted especially for ISPs and Telecom Operators.

It is written in `Go`_ programming language and is accessible from any programming language via JSON RPC. 
The code is well documented (**go doc** compliant `API docs`_) and heavily tested. (also **1300+** tests are part of the build system).

After testing various databases like `Kyoto Cabinet`_, `Apache Cassandra`_, `Redis`_ and `MongoDB`_, 
the project focused on **Redis** as it delivers the best trade-off between speed, configuration and scalability. 

.. important:: `MongoDB`_ **full** support is now added.

Thanks to CGRateS flexibility, connection to any database can be easily integrated by writing a simple adapter.

.. _CGRateS: http://cgrates.org
.. _Go: http://golang.org
.. _kyoto cabinet: http://fallabs.com/kyotocabinet
.. _apache cassandra: http://cassandra.apache.org
.. _redis: http://redis.io
.. _mongodb: http://www.mongodb.org
.. _api docs: https://godoc.org/github.com/cgrates/cgrates/apier

To better understand the CGRateS architecture, below are some logical configurations in which CGRateS can operate:

.. note::  **RALs** - is a CGRateS component and stands for RatingAccountingLCR service.

.. image::  images/Simple.png
This scenario fits most of the simple installations. The **Balancer** can be left out and the **RALs** can be queried directly.

.. image::  images/Normal.png
While the network grows more **RALs** can be thrown into the stack to offer more requests per seconds workload. 
This implies the usage of the **Balancer** to distribute the requests to the **RALs** running on the *different machines*.

.. image::  images/Normal_ha.png
Without Balancer using HA (broadcast) .... 

.. image::  images/Complicated.png
Of course more **SessionManagers** can serve *multiple Telecom Switches* and all of them are connected to the same **Balancer**. 

.. image::  images/Complicated_ha.png
Without Balancer using HA (broadcast) ....

.. note:: We are planning to support **multiple** *Balancers* for huge networks if the need arises.


1.1. CGRateS Features
---------------------

- Reliable and Fast ( very fast ;) ). To get an idea about speed, we have benchmarked 13000+ req/sec on a rather modest machine without requiring special tweaks in the kernel.
   - Using most modern programming concepts like multiprocessor support, asynchronous code execution within microthreads.
   - Built-in data caching system per call duration.
   - In-Memory database with persistence over restarts.
   - Use of Balancer assures High-Availability of RALs as well as increase of processing performance where that is required.
   - Use of Linux enterprise ready tools to assure High-Availability of the Balancer where that is required (*Supervise* for Application level availability and *LinuxHA* for Host level availability).
   - High-Availability of main components is now part of CGRateS core. 
   
- Modular architecture
    - Easy to enhance functionality by writing custom session managers or mediators.
    - Flexible API accessible via both **Gob** (Golang specific, increased performance) or **JSON** (platform independent, universally accessible).

- Prepaid, Postpaid and Pseudo-Prepaid Controller.
    - Mutiple Primary Balances per Account (eg: MONETARY, SMS, INTERNET_MINUTES, INTERNET_TRAFFIC).
    - Multiple Auxiliary Balances per Account (eg: Free Minutes per Destination,  Volume Rates, Volume Discounts).
    - Concurrent sessions per account sharing the same balance with configurable debit interval (starting with 1 second).
    - Built-in Task-Scheduler supporting both one-time as well as recurrent actions (eg: TOPUP_MINUTES_PER_DESTINATION, DEBIT_MONETARY, RESET_BALANCE).
    - ActionTriggers (useful for commercial offerings like receive amounts of monetary units if a specified number of minutes was charged in a month).

- Highly configurable Rating.
    - Connect Fees.
    - Priced Units definition.
    - Rate increments.
    - Millisecond timestaps.
    - Four decimal currencies.
    - Multiple TypeOfRecord rating (eg: standard vs. premium calls, SMSes, Internet Traffic).
    - Rating subject concatenations for combined records (eg: location based rating for same user).
    - Recurrent rates definition (per year, month, day, dayOfWeek, time).
    - Rating Profiles activation times (eg: rates becoming active at specific time in the future).

- Multi-Tenant for both Prepaid as well as Rating.

- Flexible Mediator able to run multiple mediation processes on the same CDR.

- Verbose action logging in persistent databases (eg: MongoDB/PostgreSQL/MySQL) to cope with country specific law requirements.

- Good documentation ( that's me :).

- **"Free as in Beer"** with commercial support available on-demand.


1.2. Links
----------

- CGRateS quick overview :ref:`overview-main`
- CGRateS home page `<http://www.cgrates.org>`_
- Documentation `<http://cgrates.readthedocs.io>`_
- API docs `<https://godoc.org/github.com/cgrates/cgrates/apier>`_
- Source code `<https://github.com/cgrates/cgrates>`_
- Travis CI `<https://travis-ci.org/cgrates/cgrates>`_
- Google group `<https://groups.google.com/forum/#!forum/cgrates>`_
- IRC `irc.freenode.net #cgrates <http://webchat.freenode.net/?randomnick=1&channels=#cgrates>`_

1.3. License
------------

`CGRateS`_ is released under the terms of the `[GNU GENERAL PUBLIC LICENSE Version 3] <http://www.gnu.org/licenses/gpl-3.0.en.html>`_. See **LICENSE.txt** file for details.

