.. _datadb:

DataDB
======

**DataDB** is the subsystem within **CGRateS** responsible for storing runtime data like accounts, rating plans, and other objects that the engine needs for its operation. It supports various database backends to fit different deployment needs.

Database Types
--------------

DataDB supports the following database types through the ``db_type`` parameter:

* ``*redis``: Uses Redis as the storage backend
* ``*mongo``: Uses MongoDB as the storage backend
* ``*internal``: Uses in-memory storage within the CGRateS process

When using ``*internal`` as the ``db_type``, **CGRateS** leverages your machine's memory to store all **DataDB** records directly inside the engine. This drastically increases read/write performance, as no data leaves the process, avoiding the overhead associated with external databases. The configuration supports periodic data dumps to disk to enable persistence across reboots.

The internal database is ideal for:

* Environments with extremely high performance requirements
* Systems where external database dependencies should be avoided
* Lightweight deployments or containers requiring self-contained runtime
* Temporary setups or testing environments

Remote and Replication Functionality
------------------------------------

Remote Functionality
~~~~~~~~~~~~~~~~~~~~

DataDB supports fetching data from remote CGRateS instances when items are not found locally. This allows for distributed setups where data can be stored across multiple instances.

To use remote functionality:

1. Define RPC connections to remote engines
2. Configure which data items should be fetched remotely by setting ``remote: true`` for specific items
3. Optionally set an identifier for the local connection with ``remote_conn_id``

Replication
~~~~~~~~~~~

DataDB supports replicating data changes to other CGRateS instances. When enabled, modifications (Set/Remove operations) are propagated to configured remote engines, ensuring data consistency across a distributed deployment.

Unlike the remote functionality which affects Get operations, replication applies to Set and Remove operations, pushing changes outward to other nodes.

Configuration
-------------

A complete DataDB configuration includes database connection details, remote functionality settings, replication options, and item-specific settings. For reference, the full default configuration can be found in the :ref:`configuration` section.

.. code-block:: json

    "data_db": {
        "db_type": "*redis",
        "db_host": "127.0.0.1",
        "db_port": 6379,
        "db_name": "10",
        "db_user": "cgrates",
        "db_password": "",
        "remote_conns": ["engine2", "engine3"],
        "remote_conn_id": "engine1",
        "replication_conns": ["engine2", "engine3"],
        "replication_filtered": false,
        "replication_cache": "",
        "replication_failed_dir": "/var/lib/cgrates/failed_replications",
        "replication_interval": "1s",
        "items": {
            "*accounts": {"limit": -1, "ttl": "", "static_ttl": false, "remote": false, "replicate": true},
            "*rating_plans": {"limit": -1, "ttl": "", "static_ttl": false, "remote": true, "replicate": true}
            // Other items...
        },
        "opts": {
            // Database-specific options...
        }
    }

Parameters
----------

Basic Connection
~~~~~~~~~~~~~~~~

db_type
    The database backend to use. Values: <*redis|*mongo|*internal>

db_host
    Database host address (e.g., "127.0.0.1")

db_port
    Port to reach the database (e.g., 6379 for Redis)

db_name
    Database name to connect to (e.g., "10" for Redis database number)

db_user
    Username for database authentication

db_password
    Password for database authentication

Remote Functionality
~~~~~~~~~~~~~~~~~~~~

remote_conns
    Array of connection IDs (defined in rpc_conns) that will be queried when items are not found locally

remote_conn_id
    Identifier sent to remote connections to identify this engine

Replication Parameters
~~~~~~~~~~~~~~~~~~~~~~

replication_conns
    Array of connection IDs (defined in rpc_conns) to which data will be replicated

replication_filtered
    When enabled, replication occurs only to connections that previously received a Get request for the item. Values: <true|false>

replication_cache
    Caching action to execute on replication targets when items are replicated

replication_failed_dir
    Directory to store failed batch replications when using intervals. This directory must exist before launching CGRateS.

replication_interval
    Interval between batched replications:
    - Empty/0: Immediate replication after each operation
    - Duration (e.g., "1s"): Batches replications and sends them at the specified interval

Items Configuration
~~~~~~~~~~~~~~~~~~~

DataDB manages multiple data types through the ``items`` map, with these configuration options for each item:

limit
    Maximum number of items of this type to store. -1 means no limit. Only applies to *internal database.

ttl
    Time-to-live for items before automatic removal. Empty string means no expiration. Only applies to *internal database.

static_ttl
    Controls TTL behavior. When true, TTL is fixed from initial creation. When false, TTL resets on each update. Only applies to *internal database.

remote
    When true, enables fetching this item type from remote connections if not found locally.

replicate
    When true, enables replication of this item type to configured remote connections.

Internal Database Options
~~~~~~~~~~~~~~~~~~~~~~~~~

When using ``*internal`` as the database type, additional options are available in the ``opts`` section:

internalDBDumpPath
    Defines the path to the folder where the memory-stored **DataDB** will be dumped. This path is also used for recovery during engine startup. Ensure the folder exists before launching the engine.

internalDBBackupPath
    Path where backup copies of the dump folder will be stored. Backups are triggered via the `APIerSv1.BackupDataDBDump <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.BackupDataDB>`_ API call. This API can also specify a custom path for backups, otherwise the default `internalDBBackupPath` is used. Backups serve as a fallback in case of dump file corruption or loss. The created folders are timestamped in UNIX time for easy identification of the latest backup. To recover using a backup, simply transfer the folders from a backup in internalDBBackupPath to internalDBDumpPath and start the engine. If backups are zipped, they need to be unzipped manually when restoring.

internalDBStartTimeout
    Specifies the maximum amount of time the engine will wait to recover the in-memory **DataDB** state from the dump files during startup. If this duration is exceeded, the engine will timeout and an error will be returned.

internalDBDumpInterval
    Specifies the time interval at which **DataDB** will be dumped to disk. This duration should be chosen based on the machine's capacity and data load. If the interval is set too long and a lot of data changes during that period, the dumping process will take longer, and in the event of an engine crash, any data not dumped will be lost. Conversely, if the interval is too short, and a high number of queries are done often to **DataDB**, some of the needed processing power for the queries will be used by the dump process. Since machine resources and data loads vary, it is recommended to simulate the load on your system and determine the optimal "sweet spot" for this interval. At engine shutdown, any remaining undumped data will automatically be written to disk, regardless of the interval setting.

    - Setting the interval to ``0s`` disables the periodic dumping, meaning any data in **DataDB** will be lost when the engine shuts down.
    - Setting the interval to ``-1`` enables immediate dumping—whenever a record in **DataDB** is added, changed, or removed, it will be dumped to disk immediately.
    
    Manual dumping can be triggered using the `APIerSv1.DumpDataDB <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.DumpDataDB>`_ API.

internalDBRewriteInterval
    Defines the interval for rewriting files that are not currently being used for dumping data, converting them into an optimized, streamlined version and improving recovery time. Similar to ``internalDBDumpInterval``, the rewriting will trigger based on specified intervals:

    - Setting the interval ``0s`` disables rewriting.
    - Setting the interval ``-1`` triggers rewriting only once when the engine starts.
    - Setting the interval ``-2`` triggers rewriting only once when the engine shuts down.

    Rewriting should be used sparingly, as the process temporarily loads the entire ``internalDBDumpPath`` folder into memory for optimization, and then writes it back to the dump folder once done. This results in a surge of memory usage, which could amount to the size of the dump file itself during the rewrite. As a rule of thumb, expect the engine's memory usage to approximately double while the rewrite process is running. Manual rewriting can be triggered at any time via the `APIerSv1.RewriteDataDB <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.RewriteDataDB>`_ API.

internalDBFileSizeLimit
    Specifies the maximum size a single dump file can reach. Upon reaching the limit, a new dump file is created. Limiting file size improves recovery time and allows for limit reached files to be rewritten.

Redis-Specific Options
~~~~~~~~~~~~~~~~~~~~~~

The following options in the ``opts`` section apply when using Redis:

redisMaxConns
    Connection pool size

redisConnectAttempts
    Maximum number of connection attempts

redisSentinel
    Sentinel name when using Redis Sentinel

redisCluster
    Enables Redis Cluster mode

redisClusterSync
    Sync interval for Redis Cluster

redisClusterOndownDelay
    Delay before executing commands when Redis Cluster is in CLUSTERDOWN state

redisConnectTimeout, redisReadTimeout, redisWriteTimeout
    Timeout settings for various Redis operations

redisTLS, redisClientCertificate, redisClientKey, redisCACertificate
    TLS configuration for secure Redis connections

MongoDB-Specific Options
~~~~~~~~~~~~~~~~~~~~~~~~

The following options in the ``opts`` section apply when using MongoDB:

mongoQueryTimeout
    Timeout for MongoDB queries

mongoConnScheme
    Connection scheme for MongoDB (<mongodb|mongodb+srv>)

Configuration Examples
----------------------

Persistent Internal Database
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: json

    "data_db": {
        "db_type": "*internal",
        "opts": {
            "internalDBDumpPath": "/var/lib/cgrates/internal_db/datadb",
            "internalDBBackupPath": "/var/lib/cgrates/internal_db/backup/datadb",
            "internalDBStartTimeout": "5m",
            "internalDBDumpInterval": "1m",
            "internalDBRewriteInterval": "15m",
            "internalDBFileSizeLimit": "1GB"
        }
    }

Replication Setup
~~~~~~~~~~~~~~~~~

First, define connections to the engines you want to replicate to:

.. code-block:: json

    "rpc_conns": {
        "rpl_engine": {
            "conns": [
                {
                    "address": "127.0.0.1:2012",
                    "transport": "*json",
                    "connect_attempts": 5,
                    "reconnects": -1,
                    "max_reconnect_interval": "",
                    "connect_timeout": "1s",
                    "reply_timeout": "2s"
                }
            ]
        }
    }

Then configure DataDB replication (showing only replication-related parameters):

.. code-block:: json

    "data_db": {
        "replication_conns": ["rpl_engine"],
        "replication_failed_dir": "/var/lib/cgrates/failed_replications",
        "replication_interval": "1s",
        "items": {
            "*accounts": {"replicate": true},
            "*reverse_destinations": {"replicate": true},
            "*destinations": {"replicate": true},
            "*rating_plans": {"replicate": true}
            // Other items...
        }
    }

Notes
-----

* By default, both replication and remote functionality are disabled for all items and must be explicitly enabled by setting ``replicate: true`` or ``remote: true`` for each desired item
* When using replication with intervals, make sure to configure a ``replication_failed_dir`` to handle failed replications
* Failed replications can be manually replayed using the `APIerSv1.ReplayFailedReplications <https://pkg.go.dev/github.com/cgrates/cgrates@master/apier/v1#APIerSv1.ReplayFailedReplications>`_ API call
* Remote functionality and replication can be used independently or together, depending on your deployment needs
