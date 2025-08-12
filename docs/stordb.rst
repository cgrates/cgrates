.. _stordb:

StorDB
======

**StorDB** is the subsystem within **CGRateS** responsible for storing offline data such as **tariff plans** and **CDRs**. It supports a variety of backend databases to match different scalability, performance, and deployment needs.

Database Types
--------------

StorDB supports the following database types through the ``db_type`` parameter:

* ``*mysql``: Uses MySQL as the storage backend
* ``*postgres``: Uses PostgreSQL as the storage backend
* ``*mongo``: Uses MongoDB as the storage backend
* ``*internal``: Uses in-memory storage within the CGRateS process

When using ``*internal`` as the ``db_type``, **CGRateS** stores all **StorDB** data in memory, significantly improving access speed. This type is suited for testing, lightweight deployments, and setups where storage persistence and external dependencies are not necessary.

Configuration
-------------

A complete StorDB configuration defines the storage backend, connection settings, item handling limits, and advanced item-specific options. Below is an example of a configuration using MySQL:

.. code-block:: json

    "stor_db": {
        "db_type": "*mysql",
        "db_host": "127.0.0.1",
        "db_port": 3306,
        "db_name": "cgrates",
        "db_user": "cgrates",
        "db_password": "CGRateS.org",
        "opts": {
            "sqlMaxOpenConns": 100,
            "sqlMaxIdleConns": 10,
            "sqlLogLevel": 3,
            "sqlConnMaxLifetime": "0",
            "mysqlLocation": "Local"
        }
    }

Parameters
----------

Basic Connection
~~~~~~~~~~~~~~~~

db_type
    The database backend used for storage. Supported: <*mysql|*postgres|*mongo|*internal>

db_host
    Host address for the database server

db_port
    Port to connect to the database (e.g., 3306 for MySQL)

db_name
    Database name

db_user
    Username for authenticating to the database

db_password
    Password for database authentication

string_indexed_fields
    Fields on the ``cdrs`` table to index for faster query performance (used for ``*mongo`` and ``*internal``)

prefix_indexed_fields
    Prefix-indexed fields on the ``cdrs`` table (used for ``*internal`` only)

Items Configuration
~~~~~~~~~~~~~~~~~~~

Each data type in StorDB can be configured independently under the ``items`` map:

limit
    Maximum number of items of this type to store. -1 means no limit. Only applies to *internal database.

ttl
    Time-to-live for items before automatic removal. Empty string means no expiration. Only applies to *internal database.

static_ttl
    Controls TTL behavior. When true, TTL is fixed from initial creation. When false, TTL resets on each update. Only applies to *internal database.

remote
    Not used in StorDB, included for consistency with DataDB.

replicate
    Not used in StorDB, included for consistency with DataDB.

Example:

.. code-block:: json

    "items": {
        "*cdrs": {"limit": -1, "ttl": "24h", "static_ttl": false, "remote": false, "replicate": false},
        "*tp_rates": {"limit": -1, "ttl": "1s", "static_ttl": false, "remote": false, "replicate": false}
    }

Internal Database Options
~~~~~~~~~~~~~~~~~~~~~~~~~

When ``*internal`` is selected, StorDB uses in-memory storage and supports disk persistence through the following options:

internalDBDumpPath
    Defines the path to the folder where the memory-stored **StorDB** will be dumped. This path is also used for recovery during engine startup. Ensure the folder exists before launching the engine.

internalDBBackupPath
    Path where backup copies of the dump folder will be stored. Backups are triggered via the `APIerSv1.BackupStorDBDump <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.BackupStorDB>`_ API call. This API can also specify a custom path for backups, otherwise the default `internalDBBackupPath` is used. Backups serve as a fallback in case of dump file corruption or loss. The created folders are timestamped in UNIX time for easy identification of the latest backup. To recover using a backup, simply transfer the folders from a backup in internalDBBackupPath to internalDBDumpPath and start the engine. If backups are zipped, they need to be unzipped manually when restoring.

internalDBStartTimeout
    Specifies the time interval at which **StorDB** will be dumped to disk. This duration should be chosen based on the machine's capacity and data load. If the interval is set too long and a lot of data changes during that period, the dumping process will take longer, and in the event of an engine crash, any data not dumped will be lost. Conversely, if the interval is too short, and a high number of queries are done often to **StorDB**, some of the needed processing power for the queries will be used by the dump process. Since machine resources and data loads vary, it is recommended to simulate the load on your system and determine the optimal "sweet spot" for this interval. At engine shutdown, any remaining undumped data will automatically be written to disk, regardless of the interval setting.

    - Setting the interval to ``0s`` disables the periodic dumping, meaning any data in **StorDB** will be lost when the engine shuts down.
    - Setting the interval to ``-1`` enables immediate dumping—whenever a record in **StorDB** is added, changed, or removed, it will be dumped to disk immediately.
    
    Manual dumping can be triggered using the `APIerSv1.DumpStorDB <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.DumpStorDB>`_ API.

internalDBDumpInterval
    Specifies the time interval at which **StorDB** will be dumped to disk. This duration should be chosen based on the machine's capacity and data load. If the interval is set too long and a lot of data changes during that period, the dumping process will take longer, and in the event of an engine crash, any data not dumped will be lost. Conversely, if the interval is too short, and a high number of queries are done often to **StorDB**, some of the needed processing power for the queries will be used by the dump process. Since machine resources and data loads vary, it is recommended to simulate the load on your system and determine the optimal "sweet spot" for this interval. At engine shutdown, any remaining undumped data will automatically be written to disk, regardless of the interval setting.

    - Setting the interval to ``0s`` disables the periodic dumping, meaning any data in **StorDB** will be lost when the engine shuts down.
    - Setting the interval to ``-1`` enables immediate dumping—whenever a record in **StorDB** is added, changed, or removed, it will be dumped to disk immediately.
    
    Manual dumping can be triggered using the `APIerSv1.DumpStorDB <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.DumpStorDB>`_ API.

internalDBRewriteInterval
    Defines the interval for rewriting files that are not currently being used for dumping data, converting them into an optimized, streamlined version and improving recovery time. Similar to ``internalDBDumpInterval``, the rewriting will trigger based on specified intervals:

    - Setting the interval ``0s`` disables rewriting.
    - Setting the interval ``-1`` triggers rewriting only once when the engine starts.
    - Setting the interval ``-2`` triggers rewriting only once when the engine shuts down.

    Rewriting should be used sparingly, as the process temporarily loads the entire ``internalDBDumpPath`` folder into memory for optimization, and then writes it back to the dump folder once done. This results in a surge of memory usage, which could amount to the size of the dump file itself during the rewrite. As a rule of thumb, expect the engine's memory usage to approximately double while the rewrite process is running. Manual rewriting can be triggered at any time via the `APIerSv1.RewriteStorDB <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.RewriteStorDB>`_ API.

internalDBFileSizeLimit
    Specifies the maximum size a single dump file can reach. Upon reaching the limit, a new dump file is created. Limiting file size improves recovery time and allows for limit reached files to be rewritten.

Configuration Example: Internal Storage
---------------------------------------

.. code-block:: json

    "stor_db": {
        "db_type": "*internal",
        "opts": {
            "internalDBDumpPath": "/var/lib/cgrates/internal_db/stordb",
            "internalDBBackupPath": "/var/lib/cgrates/internal_db/backup/stordb",
            "internalDBStartTimeout": "5m",
            "internalDBDumpInterval": "1m",
            "internalDBRewriteInterval": "15m",
            "internalDBFileSizeLimit": "1GB"
        }
    }


SQL-Specific Options
~~~~~~~~~~~~~~~~~~~~

sqlMaxOpenConns
    Maximum number of open connections to the database

sqlMaxIdleConns
    Maximum number of idle connections in the pool

sqlLogLevel
    Logging verbosity for SQL driver (`1` = Silent, `2` = Error, `3` = Warn, `4` = Info)

sqlConnMaxLifetime
    Maximum lifetime for a single database connection (``0`` means unlimited)

mysqlDSNParams
    Extra MySQL DSN parameters passed as a key-value object

mysqlLocation
    Timezone used by MySQL (e.g., ``Local``)


Mongo-Specific Options
~~~~~~~~~~~~~~~~~~~~

mongoQueryTimeout
    Timeout for MongoDB queries

mongoConnScheme
    Connection scheme for MongoDB: <mongodb|mongodb+srv>


Postgres-Specific Options
~~~~~~~~~~~~~~~~~~~~

sqlMaxOpenConns
    Maximum number of open connections to the database

sqlMaxIdleConns
    Maximum number of idle connections in the pool

sqlLogLevel
    Logging verbosity for SQL driver (`1` = Silent, `2` = Error, `3` = Warn, `4` = Info)

sqlConnMaxLifetime
    Maximum lifetime for a single database connection (``0`` means unlimited)

pgSSLMode
    SSL mode for PostgreSQL: <disable|allow|prefer|require|verify-ca|verify-full>. Determines whether or with what priority a secure SSL TCP/IP connection will be negotiated with the server

pgSSLCert
    File name of the client SSL certificate, replacing the default ~/.postgresql/postgresql.crt

pgSSLKey
    Location for the secret key used for the client certificate

pgSSLPassword
    Specifies the password for the secret key specified in pgSSLKey
    
pgSSLCertMode
    Determines whether a client certificate may be sent to the server, and whether the server is required to request one. <disable|allow|prefer|require|verify-ca|verify-full>

pgSSLRootCert
    Name of a file containing SSL certificate authority (CA) certificate(s)

pgSchema
    Schema name to use in PostgreSQL