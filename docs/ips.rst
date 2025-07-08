.. _ips:

IPs
===

**IPs** is the **CGRateS** component that handles IP Address Management (IPAM). It provides dynamic IP allocation and release with TTL management, filtering, and weight-based pool selection. Think of it as a smart DHCP-like service that can be integrated with billing and session management.

Processing Logic
----------------

When a request comes in, the IPs service follows this workflow:

1. **Profile Matching**: Finds the IPProfile with the highest weight that matches the event filters
2. **Caching**: Stores the matched profile ID in the ``*event_ips`` cache partition for faster subsequent operations using the same allocation ID
3. **Pool Selection**: Filters pools within the profile based on event data, sorts by weight (highest first), and stops at the first pool with blocking enabled
4. **Operation**: Performs authorize (dry-run), allocate, or release based on the API call
5. **TTL Management**: Automatically expires old allocations based on configured TTL

The service maintains allocation state across authorize, allocate, and release operations. Once a profile is matched for an allocation ID, subsequent operations use the cached profile ID for consistency and performance.

Pool Selection Logic
~~~~~~~~~~~~~~~~~~~~

Within a matched profile, pools are processed as follows:

* **Filter**: Only pools matching the event filters are considered
* **Sort**: Remaining pools are sorted by weight (highest first)
* **Block**: Processing stops at the first pool with ``Blocker: true``
* **Allocate**: Try allocation from each remaining pool until one succeeds

This allows complex allocation scenarios where high-priority pools can prevent fallback to lower-priority pools based on event conditions.

IP Allocation Strategies
------------------------

Pools are defined using CIDR notation with these allocation strategies:

.. note:: Currently supports single IP allocation per pool (CIDR /32 for IPv4, /128 for IPv6). Multi-IP allocation will be supported in future versions.

* ``*ascending``: Allocate IPs in ascending order (10.0.0.1, 10.0.0.2, ...)
* ``*descending``: Allocate IPs in descending order (10.0.0.254, 10.0.0.253, ...)
* ``*random``: Allocate IPs randomly from the range

Parameters
----------

Configure the IPs service in the **ips** section of the :ref:`JSON configuration <configuration>`:

enabled
    Enables/disables the IPs component. Values: <true|false>

store_interval
    How often to dump allocations to dataDB. Values: <""|duration>

    - "": Dump only at start/shutdown
    - ">0": Regular dump interval (e.g., "30s")
    - "<0": Never dump to DB

indexed_selects
    Enable profile matching exclusively on indexes for better performance. Values: <true|false>

string_indexed_fields
    Fields used for string-based index querying (e.g., ["*req.Account"])

prefix_indexed_fields
    Fields used for prefix-based index querying

suffix_indexed_fields
    Fields used for suffix-based index querying

exists_indexed_fields
    Fields used for existence-based index querying

notexists_indexed_fields
    Fields used for non-existence-based index querying

nested_fields
    Controls indexed filter matching depth. Values: <true|false>

    - true: checks all levels
    - false: checks only first level

opts
    Dynamic options configuration:

    *allocationID
        Defines how to extract allocation ID from events. The allocation ID uniquely identifies an IP allocation.

IPProfile
~~~~~~~~~

Defines IP allocation policies with the following parameters:

Tenant
    The tenant on the platform

ID
    Unique identifier for the profile

FilterIDs
    List of filters for profile matching

Weights
    Dynamic weights for profile selection. The profile with the highest weight that matches the event is selected.

TTL
    Time-to-live for allocations (e.g., "1h", "-1" for no expiry)

Stored
    Whether to store this profile persistently

Pools
    List of IPPool objects defining available IP ranges

IPPool
~~~~~~

Defines individual IP pools within a profile:

ID
    Unique identifier for the pool

FilterIDs
    List of filters for pool matching

Type
    Pool type (*ipv4 or *ipv6)

Range
    IP range in CIDR notation (e.g., "192.168.1.0/24")

Strategy
    Allocation strategy (*ascending, *descending, *random)

Message
    Custom message returned with allocated IP

Weights
    Dynamic weights for pool selection. Higher weight pools are tried first.

Blockers
    Dynamic blockers that stop pool processing when true

API Methods
-----------

V1AuthorizeIP
~~~~~~~~~~~~~

Checks if an IP can be allocated without actually allocating it (dry run).

**Request:**

::

   {
     "Tenant": "cgrates.org",
     "ID": "unique_event_id",
     "Event": {
       "Account": "1001",
       "Destination": "1002"
     },
     "APIOpts": {
       "*ipAllocationID": "ip_allocation_abc123"
     }
   }

**Returns:**

AllocatedIP object with the following fields:

- ProfileID: ID of the matched IPProfile
- PoolID: ID of the selected pool
- Message: Custom message from the pool configuration  
- Address: IP address that would be allocated

**Example Response:**

::

   {
     "ProfileID": "IPsAPI",
     "PoolID": "POOL_C",
     "Message": "Pool C message",
     "Address": "10.100.0.3"
   }

Returns error if allocation would fail.

V1AllocateIP
~~~~~~~~~~~~

Allocates an IP address for the event. If the allocation ID already exists, refreshes the allocation timestamp.

**Request:** Same format as V1AuthorizeIP

**Returns:**

AllocatedIP object with allocated IP details (same format as V1AuthorizeIP). Returns error if allocation fails.

V1ReleaseIP
~~~~~~~~~~~

Releases a previously allocated IP address.

**Request:** Same format as V1AuthorizeIP

**Returns:**

- Success confirmation
- Error if release fails

V1GetIPAllocations
~~~~~~~~~~~~~~~~~~

Retrieves current allocation state for a profile.

**Parameters:**

- Tenant and Profile ID

**Returns:**

- IPAllocations object with current allocation state

V1GetIPAllocationForEvent
~~~~~~~~~~~~~~~~~~~~~~~~~

Gets the matching IPAllocations object for a specific event.

**Parameters:**

- Event with allocation ID

**Returns:**

- IPAllocations object for the matching profile

Use Cases
---------

- **RADIUS Integration**: Assign Framed-IP-Address for network sessions
- **Temporary Allocations**: IP allocation for time-limited connections or services
- **Multi-tenant IPAM**: Separate IP pools per tenant/customer

Example Configuration
---------------------

::

   {
     "ips": {
       "enabled": true,
       "store_interval": "30s",
       "indexed_selects": true,
       "string_indexed_fields": ["*req.Account"],
       "opts": {
         "*allocationID": [
           {
             "Tenant": "cgrates.org",
             "FilterIDs": ["*string:~*req.Account:1001"],
             "Value": "ip_session_fixed_id"
           }
         ]
       }
     }
   }

Example IPProfile
-----------------

::

   {
     "Tenant": "cgrates.org",
     "ID": "CUSTOMER_POOL",
     "FilterIDs": ["*string:~*req.Account:1001"],
     "Weights": [{"Weight": 10}],
     "TTL": "24h",
     "Pools": [
       {
         "ID": "PREMIUM_POOL",
         "Type": "*ipv4",
         "Range": "10.1.1.0/24",
         "Strategy": "*ascending",
         "Message": "Premium IP allocated",
         "Weights": [{"Weight": 100}],
         "FilterIDs": ["*string:~*req.Plan:premium"]
       },
       {
         "ID": "STANDARD_POOL",
         "Type": "*ipv4",
         "Range": "10.1.2.0/24",
         "Strategy": "*ascending",
         "Message": "Standard IP allocated",
         "Weights": [{"Weight": 50}]
       }
     ]
   }

In this example, premium plan users get IPs from the premium pool, while others get standard pool IPs. If the premium pool is exhausted, premium users fall back to the standard pool.
