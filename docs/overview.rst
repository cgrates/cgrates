.. _overview-main:

1. Overview
===========
Starting as a pure **billing engine**, CGRateS has evolved over the years into a reliable **real-time charging framework** able to accommodate various business cases in a *generic way*. 
Meant to be pluggable into existing billing infrastructure and as non-intrusive as possible, 
CGRateS passes the decisions about logic flow to system administrators and incorporates as less as possible business logic.

Being an *"engine style"* the project focuses on providing best ratio between **functionality** (
over 15 daemons/services implemented, 
Multi-tenancy, 
derived charging - eg: chaining of the business resellers, 
account bundles, 
LCR, 
CDRStatS, 
Diameter Server, 
A-Number rating, 
built-in High-Availability support
agile in developing new features 
) 
and **performance** (
dedicated benchmark tool, 
asynchronous request processing, 
own transactional cache with majority of handled data loaded on start or reloaded during runtime, 
built-in balancer
) 
however not loosing focus of **quality** (over 1300 tests part of the build environment).

Modular and flexible, CGRateS provides APIs over a variety of simultaneously accessible communication interfaces:
 - **In-process**           : optimal when there is no need to split services over different processes.
 - **JSON over TCP**        : most preferred due to its simplicity and readability. 
 - **JSON over HTTP**       : popular due to fast interoperability development.
 - **JSON over Websockets** : useful where 2 ways interaction over same TCP socket is required.
 - **GOB over TCP**         : slightly faster than JSON one but only accessible for the moment out of Go (`<https://golang.org/>`_).

CGRateS is capable of four charging modes

- \*prepaid
   - Session events monitored in real-time
   - Session authorization via events with security call timer
   - Real-time balance updates with configurable debit interval
   - Support for simultaneous sessions out of the same account
   - Real-time fraud detection with automatic mitigation

- \*pseudoprepaid
   - Session authorization via events
   - Charging done at the end of the session out of CDR received
   - Advantage: less CPU intensive due to less events processed
   - Disadvantage: as balance updates happen only at the end of the session there can be costs discrepancy in case of multiple sessions out of same account 
     (including going on negative balance).

- \*postpaid
   - Charging done at the end of the session out of CDR received without session authorization
   - Useful when no authorization is necessary (trusted accounts) and no real-time event interaction is present (balance is updated only when CDR is present).

- \*rated
   - Special charging mode where there is no accounting interaction (no balances are used) but the primary interest is attaching costs to CDRs.
   - Specific mode for Wholesale business processing high-throughput CDRs.
   - Least CPU usage out of the four modes (fastest charging)

2. CGRateS Subsystems
=====================


2.1. RALs (RatingAccountingLCRservice)
--------------------------------------
- Primary component, offering the most functionality out of the subsystems.
- Computes replies based on static list of "rules" defined in TariffPlan.

2.1.1. Rater 
~~~~~~~~~~~~
- Defines the performance of the system as a whole being the "heart" component
- Support for multiple TypeOfRecord (**\*voice**, **\*data**, **\*sms**, **\*generic**)
- Time based calculations (activation time in the future/rate-destination timely coupled) with granular time definitions (year, month, month day, weekday, time in seconds)
- Compressed destination prefixes, helping on faster destination match as well as memory consumption
- Advanced Rating capabilities: 
  ConnectFee (charged at beginning of the session); 
  RateUnit (automatic divider for the cost); 
  RateIncrement (increase verbosity of the charging interval); 
  Grouped interval rating inside the call duration (charging each second within a session independently)
- Per destination rounding: control number of decimals displayed in costs, decide rounding methods (**\*up**, **\*down**, **\*middle**)
- Control of the MaxSessionCost with decision on action taken on threshold hit (**\*free**, **\*disconnect**)
- Unlimited chaining of rating profiles (escalation price lists)

2.1.2. Accounting
~~~~~~~~~~~~~~~~~
- Maintains accounts with bundles and usage counters
- Support for multiple TypeOfRecord (**\*voice**, **\*data**, **\*sms**, **\*generic**)
- Unlimited number of balances per account
- Balance prioritization via balance weights
- Advanced balance selection (Direction, Destinations, RatingSubject - volume discounts in real-time, Categories)
- Accurate balance lifespan definition (ExpirationDate, Activation intervals)
- Safe account operations via in-/inter-process locks and on-disk storage
- Shared balances between multiple accounts (family/company bundles) with per-consumer configurable debit strategy and rates selected.
- Concurrent sessions per account doing balance reservation in chunks of debit interval and support for refunds and debit sleep when needed
- Scheduled account operations via predefined actions (eg: **\*topup**, **\*debit**) or notifications (**\*http_call_url**, **\*mail**)
- Fraud detection with automatic mitigation via action triggers/thresholds monitoring both balance status as well as combined usage

2.1.3. LCR
~~~~~~~~~~
- Accessible via RPC for queries or coupled with external communication systems sharing supplier information via specific channel variables.
- Integrates traffic patterns (LCR for specific session duration)
- Advanced profile selection mechanism (Direction, Tenant, Category, Account, Subject, Destination).
- Weight based prioritisation.
- Profile activation in the future possible through ActivationTime parameter.
- Tightly coupled with Accounting subsystem providing LCR over bundles (eg: consider minutes with special price only during weekend)
- Extended functionality through the use of strategies and individual parameters per strategy
   - **\*static**: list of suppliers is always statically returned, independent on cost
   - **\*least_cost**: classic LCR where suppliers are ordered based on cheapest cost
   - **\*highest_cost**: suppliers are ordered based on highest cost
   - **\*qos_thresholds**: suppliers are ordered based on cheapest cost and considered only if their quality stats (ASR, ACD, TCD, ACC, TCC, PDD, DDC) are within the defined intervals
   - **\*qos**: suppliers are ordered by their quality stats (ASR, ACD, TCD, ACC, TCC, PDD, DDC)
   - **\*load_distribution**: suppliers are ordered based on preconfigured load distribution scheme, independent on their costs.

2.2. CDRs
---------
- Real-time, centralized CDR server designed to receive CDRs via RPC interfaces
- Attaches Costs received from RALs to CDR events
- Offline CDR storage
- Real-time CDR replication to multiple upstream servers (CDR Rating queues) for high performance (optional disk-less) CDR processing
- Flexible export interfaces (JSON templates) with output mediation
- SureTax integration for US specific tax calculations

2.3. CDRStatS
-------------
- Compute real-time stats based on CDR events received
- In-memory / performance oriented
- Unlimited StatQueues computing the same CDR event
- Flexible queue configuration (QueueLength, TimeWindow, Metrics, CDR field filters)
- Fraud detection with automatic mitigation through action triggers 

2.4. AliaseS
------------
- Context based data aliasing (**\*rating** - converts data on input before calculations)
- Multiple layers for filtering (Direction, Tenant, Category, Account, Subject, DestinationID, Context)
- Multiple fields replaced simultaneously based on Target parameter

2.5. UserS
----------
- Populate requests with user profile fields (replace **\*users** marked fields with data from matched profile)
- Best match inside user properties
- Attribute-value store (similar to LDAP/Diameter)

2.6. RLs (ResourceLimiterService)
---------------------------------
- Limits resources during authorization (eg: maximum calls per destination for an account)
- Time aware (resources available during predefined time interval)

2.7. PubsubS
------------
- Expose internal events to subscribed external entities (eg: real-time balance updates being sent to an external http server)
- Advanced regexp filters for subscriptions
- Configurable subscription lifespan

2.8. HistoryS
-------------
- Archive rate changes in git powered environment
- In-memory diffs with regular dumps to filesystem

2.9. DA (DiameterAgent)
-----------------------
- Diameter **server** implementation
- Flexible processing logic configured inside JSON templates (standard agnostic)
- Mediation for incoming fields (regexp support with in-memory compiled rules).

2.10. SM (SessionManager)
-------------------------
- Maintain/disconnect sessions
- Balance reservation and refunds

2.10.1. SMG (SessionManagerGeneric)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
- Switch agnostic session management via RPC interface
- Bi-JSONRPC support

2.10.2. SMG-Asterisk
~~~~~~~~~~~~~~~~~~~~
- Asterisk specific communication over ARI and AMI interfaces
- Bidirectional (subscribing for events as well as sending commands)

2.10.3. SM-FreeSWITCH
~~~~~~~~~~~~~~~~~~~~~
- FreeSWITCH specific communication interface via ESL
- Bidirectional (subscribing for events as well as sending commands)
- Zero configuration in FreeSWITCH for CDR generation (useful for billing assurance/parallel billing)
- Ability to manage multiple FreeSWITCH servers from the same CGR-SM component

2.10.4. SM-Kamailio
~~~~~~~~~~~~~~~~~~~
- Bidirectional Kamailio communication via evapi
- Ability to manage multiple Kamailio instances from the same CGR-SM component

2.10.5. SM-OpenSIPS
~~~~~~~~~~~~~~~~~~~
- Bidirectional OpenSIPS communication via event_diagram/mi_datagram
- Deadlink detection via subscription mechanism

2.11. CDRC (CDR Client)
-----------------------
- Offline CDR processing for **.csv**, **.xml** and **.fwv** file sources
- Mediation via in-memory regexp rules inside JSON templates
- Linux inotify support for instant file processing or delayed folder monitoring


3. CGRateS Peripherals
======================
Packaged together due to common usage
 
3.1. cgr-engine
---------------
- Configured via .json files, encorporating CGRateS subsystems mentioned above
- Can start as many / less services as needed communicating over internal or external sockets
- Multiple cgr-engine processes can be started on the same host
- Asynchronous service runs (services synchronize later inside process via specific communication channels, however they all run independent of each other).
- RPC Server with multiple interfaces started automatically based on needs.
- TCP sockets shared between services

3.2. cgr-console
----------------
- Application interfacing with cgr-engine via TCP sockets (JSON serialization)
- History and help command support

3.3. cgr-loader
---------------
- Loads TariffPlan data out of .csv files into CGRateS live database or imports it into offline one for offline management
- Automatic cache reloads with optimizations for data loaded

3.4. cgr-tester
---------------
- Benchmarking tool to test based on particular TariffPlans of users.

3.5. cgr-admin (`<https://github.com/cgrates/cgradmin>`_)
----------------------------------------------------
- PoC web interface demonstrating recommended way to interact with CGRateS from an external GUI.

4. Fraud detection within CGRateS
=================================
- Due to its importance in billing, CGRateS has invested considerable efforts into fraud detection and automatic mitigation.
- For redundancy and reliability purposes, there are two mechanisms available within CGRateS to detect fraud.

4.1. Fraud detection within Accounting:
---------------------------------------
- Events are happening in real-time, being available during updates (eg: every n seconds of a session).
- Thresholds set by the administrator are reacting by calling a set of predefined actions **synchronously** 
  (with the advantage of having account in locked state, eg. no other events are possible until decision is made) or **asynchronously** (unlocking the accounts faster)
- Two types of thresholds can be set 
   - **min-/max-balance** monitoring balance values 
   - **min-/max-usage** counters (eg: amount of minutes to specific destination).
- Middle session control (sessions can be disconnected as fraud is detected

4.2. Fraud detection within CDRStatS:
-------------------------------------
- Thresholds are monitoring CDRStatS queues and reacting by calling synchronously or asynchronously a set of predefined actions.
- Various stats metrics can be monitored (min-/max- ASR, ACD, TCD, ACC, TCC, PDD, DDC)

