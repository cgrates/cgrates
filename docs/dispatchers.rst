.. _dispatchers:

DispatcherS
===========

**DispatcherS** is the **CGRateS** component that handles request routing and load balancing. When enabled, it manages all requests to other CGRateS subsystems by wrapping their methods with additional features like :ref:`authorization <dispatcher-authorization>`, request routing, load balancing, broadcasting, and loop prevention.

Processing Logic
----------------

.. _dispatcher-authorization:

Authorization
~~~~~~~~~~~~~

Optional step when :ref:`AttributeS <attributes>` connections are configured:

- Looks for ``*apiKey`` in APIOpts (returns mandatory missing error if not present)
- Sends key to AttributeS, which adds it as "APIKey" to the Event
- AttributeS processes the event and adds allowed methods to APIMethods (e.g., "method1&method2&method3")
- Checks if the requested method is in the allowed list
- Continues only after successful authorization

Dispatch
~~~~~~~~

The dispatcher processes requests through these steps:

* Check for bypass conditions:
   * Presence of ``*dispatchers: false`` in APIOpts
   * Request source is another dispatcher and ``prevent_loops`` is enabled

* Check cached routes:
   * Search for ``*routeID`` in APIOpts
   * If found, use cached dispatch data (tenant, profile ID, host ID)
   * Fall back to full dispatch on network errors or timeouts

* Run full dispatch sequence:
   * Get matching dispatcher profiles
   * Try each profile until dispatch succeeds

.. _dispatcher-types:

Dispatcher Types and Strategies
-------------------------------

Load Dispatchers
~~~~~~~~~~~~~~~~

Used for ratio-based request distribution. Hosts are sorted in three steps:

1. Initial sorting by weight
2. Secondary sorting by load ratio (current active requests/configured ratio), where lower ratios have priority
3. Final sorting based on the specified strategy:

   * ``*random``: Randomizes host selection 
   * ``*round_robin``: Sequential host selection with weight consideration
   * ``*weight``: Skips final sorting, maintains weight and load-based ordering

Configuration through:

- ``*defaultRatio`` in StrategyParams
- Direct ratio specification in Host configuration


Simple Dispatchers
~~~~~~~~~~~~~~~~~~

Standard request distribution where hosts are sorted first by weight, followed by the chosen strategy (*random, *round_robin, *weight).

Broadcast Dispatchers
~~~~~~~~~~~~~~~~~~~~~

Handles scenarios requiring multi-host distribution. Supports three broadcast strategies:

* ``*broadcast``: Sends to all hosts, uses first response
* ``*broadcast_sync``: Sends to all hosts, waits for all responses
* ``*broadcast_async``: Sends to all hosts without waiting (fire-and-forget)

Parameters
----------

Configure the dispatcher in the **dispatchers** section of the :ref:`JSON configuration <configuration>`:

enabled
    Enables/disables the DispatcherS component. Values: <true|false>

indexed_selects
    Enables profile matching exclusively on indexes for improved performance

string_indexed_fields
    Fields used for string-based index querying

prefix_indexed_fields
    Fields used for prefix-based index querying

suffix_indexed_fields
    Fields used for suffix-based index querying

nested_fields
    Controls indexed filter matching depth. Values: <true|false>
    - true: checks all levels
    - false: checks only first level

attributes_conns
    Connections to :ref:`AttributeS <attributes>` for API authorization
    - Empty: disables authorization
    - "*internal": uses internal connection
    - Custom connection ID

any_subsystem
    Enables matching of *any subsystem. Values: <true|false>

prevent_loops
    Prevents request loops between dispatcher nodes. Values: <true|false>

DispatcherHost
~~~~~~~~~~~~~~

Defines individual dispatch destinations with the following parameters:

Tenant
    The tenant on the platform

ID
    Unique identifier for the host

Address
    Host address (use *internal for internal connections)

Transport
    Protocol used for communication (*gob, *json)

ConnectAttempts
    Number of connection attempts

Reconnects
    Maximum number of reconnection attempts

MaxReconnectInterval
    Maximum interval between reconnection attempts

ConnectTimeout
    Connection timeout (e.g., "1m")

ReplyTimeout
    Response timeout (e.g., "2m")

TLS
    TLS connection settings:
    - ClientKey: Path to client key file
    - ClientCertificate: Path to client certificate
    - CaCertificate: Path to CA certificate

DispatcherProfile
~~~~~~~~~~~~~~~~~

Defines routing rules and strategies. See :ref:`dispatcher-types` for available strategies.

Tenant
    The tenant on the platform

ID
    Profile identifier

Subsystems
    Target subsystems (*any for all)

FilterIDs
    List of filters for request matching

ActivationInterval
    Time interval when profile is active

Strategy
    Dispatch strategy (*weight, *random, *round_robin, *broadcast, *broadcast_sync)

StrategyParameters
    Additional strategy configuration (e.g., *default_ratio)

ConnID
    Target host identifier

ConnFilterIDs
    Filters for connection selection

ConnWeight
    Priority weight for connection selection within the profile

ConnBlocker
    Blocks connection if true

ConnParameters
    Additional connection parameters (e.g., *ratio)

Weight
    Priority weight used when selecting between multiple matching profiles

Use Cases
---------

- Load balancing between multiple CGRateS nodes
- High availability setups with automatic failover
- Request authorization and access control
- Broadcasting requests for data collection
- Traffic distribution based on weight or custom metrics
- System scaling and performance optimization
