.. _CGRateS: http://cgrates.org
.. _Go: http://golang.org
.. _Docker: https://www.docker.com/
.. _Kafka: https://kafka.apache.org/
.. _redis: http://redis.io
.. _mongodb: http://www.mongodb.org
.. _api docs: https://godoc.org/github.com/cgrates/cgrates/apier
.. _SQS: https://aws.amazon.com/de/sqs/
.. _AMQP: https://www.amqp.org/
.. _Asterisk: https://www.asterisk.org/
.. _FreeSWITCH: https://freeswitch.com/
.. _Kamailio: https://www.kamailio.org/w/
.. _OpenSIPS: https://opensips.org/


Introduction
============


`CGRateS`_ is a *very fast* (**50k+ CPS**) and *easily scalable* (**load-balancer** + **replication** included) **Real-time Enterprise Billing Suite** targeted especially for ISPs and Telecom Operators (but not only).

Starting as a pure **billing engine**, CGRateS has evolved over the years into a reliable **real-time charging framework**, able to accommodate various business cases in a *generic way*.

Being an *"engine style"* the project focuses on providing best ratio between **functionality** (over 15 daemons/services implemented with a rich number of `features <cgrates_features>`_ and a development team agile in developing new ones) and **performance** (dedicated benchmark tool, asynchronous request processing, own transactional cache component), however not losing focus of **quality** (test driven development policy).

It is written in `Go`_ programming language and accessible from any programming language via JSON RPC.
The code is well documented (**go doc** compliant `API docs`_) and heavily tested (**5k+** tests are part of the unit test suite).

Meant to be pluggable into existing billing infrastructure and as non-intrusive as possible,
CGRateS passes the decisions about logic flow to system administrators and incorporates as less as possible business logic.

Modular and flexible, CGRateS provides APIs over a variety of simultaneously accessible communication interfaces:
 - **In-process**           : optimal when there is no need to split services over different processes
 - **JSON over TCP**        : most preferred due to its simplicity and readability
 - **JSON over HTTP**       : popular due to fast interoperability development
 - **JSON over Websockets** : useful where 2 ways interaction over same TCP socket is required
 - **GOB over TCP**         : slightly faster than JSON one but only accessible for the moment out of Go (`<https://golang.org/>`_).

CGRateS is capable of four charging modes:

- \*prepaid
   - Session events monitored in real-time
   - Session authorization via events with security call timer
   - Real-time balance updates with configurable debit interval
   - Support for simultaneous sessions out of the same account
   - Real-time fraud detection with automatic mitigation
   - *Advantage*: real-time overview of the costs and fast detection in case of fraud, concurrent account sessions supported 
   - *Disadvantage*: more CPU intensive.

- \*pseudoprepaid
   - Session authorization via events
   - Charging done at the end of the session out of CDR received
   - *Advantage*: less CPU intensive due to less events processed
   - *Disadvantage*: as balance updates happen only at the end of the session there can be costs discrepancy in case of multiple sessions out of same account (including going on negative balance).

- \*postpaid
   - Charging done at the end of the session out of CDR received without session authorization
   - Useful when no authorization is necessary (trusted accounts) and no real-time event interaction is present (balance is updated only when CDR is present).

- \*rated
   - Special charging mode where there is no accounting interaction (no balances are used) but the primary interest is attaching costs to CDRs.
   - Specific mode for Wholesale business processing high-throughput CDRs
   - Least CPU usage out of the four modes (fastest charging).


More overview content:

.. toctree::
   :maxdepth: 1

   overview.rst


Table of Contents
-----------------

.. toctree::
   :maxdepth: 5
   
   architecture
   installation
   configuration
   administration
   advanced
   tutorials
   miscellaneous








