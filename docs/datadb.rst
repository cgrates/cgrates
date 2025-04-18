.. _datadb:

DataDB
======

**DataDB** is the subsystem within **CGRateS** responsible for storing internal engine data, supporting various databases such as Redis, Mongo, or the high-performance in-memory option: `*internal`.

When using `*internal` as the `db_type`, **CGRateS** leverages your machine’s memory to store all **DataDB** records directly inside the engine. This drastically increases read/write performance, as no data leaves the process, avoiding the overhead associated with external databases like Redis or Mongo.

The `*internal` option is especially suitable for high-throughput environments, allowing **CGRateS** to operate at peak speed when accessing or modifying stored records. Additionally, this configuration supports periodic data dumps to disk to enable persistence across reboots.

Configuration example
---------------------

A configuration for using `*internal` as **DataDB** looks as the following:

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

Parameters
----------

\internalDBDumpPath
    Defines the path to the folder where the memory-stored **DataDB** will be dumped. This path is also used for recovery during engine startup. Ensure the folder exists before launching the engine.

\internalDBBackupPath
    Path where backup copies of the dump folder will be stored. Backups are triggered via the `APIerSv1.BackupDataDBDump <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.BackupDataDB>`_ API call. This API can also specify a custom path for backups, otherwise the default `internalDBBackupPath` is used. Backups serve as a fallback in case of dump file corruption or loss. The created folders are timestamped in UNIX time for easy identification of the latest backup. To recover using a backup, simply transfer the folders from a backup in internalDBBackupPath to internalDBDumpPath and start the engine. If backups are zipped, they need to be unzipped manually when restoring.

\internalDBStartTimeout
    Specifies the maximum amount of time the engine will wait to recover the in-memory **DataDB** state from the dump files during startup. If this duration is exceeded, the engine will timeout and an error will be returned.

\internalDBDumpInterval
    Specifies the time interval at which **DataDB** will be dumped to disk. This duration should be chosen based on the machine's capacity and data load. If the interval is set too long and a lot of data changes during that period, the dumping process will take longer, and in the event of an engine crash, any data not dumped will be lost. Conversely, if the interval is too short, and a high number of queries are done often to **DataDB**, some of the needed processing power for the queries will be used by the dump process. Since machine resources and data loads vary, it is recommended to simulate the load on your system and determine the optimal "sweet spot" for this interval. At engine shutdown, any remaining undumped data will automatically be written to disk, regardless of the interval setting.

- Setting the interval to `0s` disables the periodic dumping, meaning any data in **DataDB** will be lost when the engine shuts down.
- Setting the interval to `-1` enables immediate dumping—whenever a record in **DataDB** is added, changed, or removed, it will be dumped to disk immediately.
Manual dumping can be triggered using the `APIerSv1.DumpDataDB <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.DumpDataDB>`_ API.

\internalDBRewriteInterval
    Defines the interval for rewriting files that are not currently being used for dumping data, converting them into an optimized, streamlined version and improving recovery time. Similar to `internalDBDumpInterval`, the rewriting will trigger based on specified intervals:

- Setting the interval `0s` disables rewriting.
- Setting the interval `-1` triggers rewriting only once when the engine starts.
- Setting the interval `-2` triggers rewriting only once when the engine shuts down.

Rewriting should be used sparingly, as the process temporarily loads the entire `internalDBDumpPath` folder into memory for optimization, and then writes it back to the dump folder once done. This results in a surge of memory usage, which could amount to the size of the dump file itself during the rewrite. As a rule of thumb, expect the engine's memory usage to approximately double while the rewrite process is running. Manual rewriting can be triggered at any time via the `APIerSv1.RewriteDataDB <https://pkg.go.dev/github.com/cgrates/cgrates@master/engine#InternalDB.RewriteDataDB>`_ API.

\internalDBFileSizeLimit
    Specifies the maximum size a single dump file can reach. Upon reaching the limit, a new dump file is created. Limiting file size improves recovery time and allows for limit reached files to be rewritten.

Use cases
---------

* Deploying **CGRateS** in environments with extremely high read/write performance requirements.
* Systems where external database dependencies are undesired or unavailable.
* Lightweight deployments or containers requiring a self-contained runtime.
* Scenarios requiring minimal latency for internal data access.
* Temporary setups or testing environments that can leverage memory-based persistence.
