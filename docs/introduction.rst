************
Introduction
************

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


.. _cgrates_features:

Features
========

- Performance oriented. To get an idea about speed, we have benchmarked 50000+ req/sec on comodity hardware without any tweaks in the kernel
    - Using most modern programming concepts like multiprocessor support, asynchronous code execution within microthreads, channel based locking
    - Built-in data caching system with LRU and TTL support
    - Linear performance increase via simple hardware addition
    - On demand performance increase via in-process / over network communication between engine services. 

- Modular architecture
    - Plugable into existing infrastructure
    - Non-intrusive into existing setups
    - Easy to enhance functionality by writing custom components
    - Flexible API accessible via both **GOB** (`Go`_ specific, increased performance) or **JSON** (platform independent, universally accessible)
    - Easy distribution (one binary concept, can run via NFS on all Linux servers without install).

- Easy administration
    - One binary can run on all Linux servers without additional installation (simple copy)
    - Can run diskless via NFS
    - Virtualization/containerization friendly(runs on Docker_).

- GOCS (Global Online Charging System)
    - Support for global networks with one master + multi-cache nodes around the globe for low query latency
    - Mutiple Balance types per Account (\*monetary, \*voice, \*sms, \*data, \*generic)
    - Unlimited number of Account Balances with weight based prioritization
    - Various Balance filters (ie: per-destination, roaming-only, weekend-only)
    - Support for Volume based discounts and automatic bonuses (ie: 5 SMS free for every 10 minutes in one hour to specific destination)
    - Session based charging with support for concurrent sessions per account and per session dynamic debit interval
    - Session emulation combined with Derived Charging (separate charging for distributors chaining, customer/supplier parallel calculations)
    - Balance reservation and refunds
    - Event based charging (ie: SMS, MESSAGE)
    - Built-in Task-Scheduler supporting both one-time as well as recurrent actions (automatic subscriptions management, recurrent \*debit/\*topup, DID charging)
    - Real-time balance monitors with automatic actions triggered (bonuses or fraud detection).

- Highly configurable Rating
    - Connect Fees
    - Priced Units definition
    - Rate increments
    - Rate groups (ie. charge first minute in a call as a whole and next ones per second)
    - Verbose durations(up to nanoseconds billing)
    - Configurable decimals per destination
    - Rating subject categorization (ie. premium/local charges, roaming)
    - Recurrent rates definition (per year, month, day, dayOfWeek, time)
    - Rating Profiles activation times (eg: rates becoming active at specific time in the future)
    - Rating Profiles fallback (per subject destinations with fallback to server wide pricing)
    - Verbose charging logs to comply strict rules imposed by some country laws.

- Multi-Tenant from day one
    - Default Tenant configurable for one-tenant systems
    - Security enforced for RPC-API on Tenant level.

- Online configuration reloads without restart
    - Engine configuration from .json folder or remote http server
    - Tariff Plans from .csv folder or database storage.

- CDR server
    - Optional offline database storage
    - Online (rating queues) or offline (via RPC-API) exports with customizable content via .json templates
    - Multiple export interfaces: files, HTTP, AMQP_, SQS_, Kafka_.

- Generic Event Reader
    - Process various sources of events and convert them into internal ones which are sent to CDR server for rating
    - Conversion rules defined in .json templates
    - Supported interfaces: .csv, .xml, fixed width files, Kafka_.

- Events mediation
    - Ability to add/change/remove information within *Events* to achieve additional services or correction
    - Performance oriented.

- Routing server for VoIP
    - Implements strategies like *Least Cost Routing*, *Load Balacer*, *High Availability*
    - Implements *Number Portability* service.

- Resource allocation controller
    - Generic filters for advanced logic
    - In-memory operations for increased performance
    - Backup in offline storage.

- Stats service
    - Generic stats (\*sum, \*difference, \*multiply, \*divide)
    - In-memory operations for increased performance
    - Backup in offline storage.

- Thresholds monitor
    - Particular implementation of *Fraud Detection with automatic mitigation*
    - Execute independent actions which can serve various purposes (notifications, accounts disables, bonuses to accounts).

- Multiple RPC interfaces
    - Support for *JSON-RPC*, *GOB-PC* over TCP, HTTP, websockets
    - Support for HTTP-REST interface.

- Various agents to outside world:
    - Asterisk_
    - FreeSWITCH_
    - Kamailio_
    - OpenSIPS_
    - Diameter
    - Radius
    - Generic HTTP
    - DNS/ENUM.

- Built in High-Availability mechanisms:
    - Dispatcher with static or dynamic routing
    - Server data replication
    - Client remote data querying.


- Good documentation ( that's me :).

- **"Free as in Beer"** with commercial support available on-demand.


Links
=====

- CGRateS home page `<http://www.cgrates.org>`_
- Documentation `<http://cgrates.readthedocs.io>`_
- API docs `<https://godoc.org/github.com/cgrates/cgrates/apier>`_
- Source code `<https://github.com/cgrates/cgrates>`_
- Travis CI `<https://travis-ci.org/cgrates/cgrates>`_
- Google group `<https://groups.google.com/forum/#!forum/cgrates>`_
- IRC `irc.freenode.net #cgrates <http://webchat.freenode.net/?randomnick=1&channels=#cgrates>`_


License
=======

`CGRateS`_ is released under the terms of the `[GNU GENERAL PUBLIC LICENSE Version 3] <http://www.gnu.org/licenses/gpl-3.0.en.html>`_. See **LICENSE.txt** file for details.


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


