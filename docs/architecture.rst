2. Architecture
===============
The CGRateS suite consists of **five** software applications described below.

.. hlist::
   :columns: 5

   - cgr-engine
   - cgr-loader
   - cgr-console
   - cgr-tester
   - cgr-migrator


CGRateS has an internal cache.

::

   "internal_cache" - cache

Operates with different external databases mentioned below.

::

   "data_db"       - MongoDB, Redis
   "stor_db"       - MongoDB, MySQL, PostgreSQL


.. hlist::
   :columns: 1

   - **data_db**       - used to store runtime data ( eg: accounts )
   - **stor_db**       - used to store offline tariff plan(s) and CDRs


.. figure::  images/CGRateSInternalArchitecture.png
   :alt: CGRateS Internal Architecture
   :align: Center
   :scale: 75 %


   CGRateS high level design

2.1. cgr-engine
---------------
Is the most important and complex component.
Customisable through the use of *json* configuration file(s),
it will start on demand **one or more** service(s), outlined below.

::

 cgrates@OCS:~$ cgr-engine -help
 Usage of cgr-engine:
   -cdrs
         Enforce starting of the cdrs daemon overwriting config
   -config_path string
         Configuration directory path. (default "/etc/cgrates/")
   -cpuprofile string
         write cpu profile to file
   -pid string
         Write pid file
   -rater
         Enforce starting of the rater daemon overwriting config
   -scheduler
         Enforce starting of the scheduler daemon .overwriting config
  -scheduled_shutdown string
         shutdown the engine after this duration
   -singlecpu
         Run on single CPU core
   -version
         Prints the application version.


.. hint::  # cgr-engine -config_path=/etc/cgrates


2.1.1. RALs service
~~~~~~~~~~~~~~~~~~~~
Responsible with the following tasks:

   - Operates on balances.
   - Computes prices for rating subjects.
   - Monitors and executes triggers.
   - LCR functionality

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   "data_db"       - (dataDb)
   "stor_db"       - (cdrDb, loadDb)

- Config section in the CGRateS configuration file:
   - ``"rals": {...}``

2.1.2. Scheduler service
~~~~~~~~~~~~~~~~~~~~~~~~
Used to execute periodic/scheduled tasks.

- Communicates via:
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   "data_db" - (dataDb)

- Config section in the CGRateS configuration file:
   - ``"scheduler": {...}``

2.1.3. SessionManager service
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Responsible with call control on the Telecommunication Switch side. Operates in two different modes (per call or globally):

- PREPAID
   - Monitors call start.
   - Checks balance availability for the call.
   - Enforces global timer for a call at call-start.
   - Executes routing commands for the call where that is necessary ( eg call un-park in case of FreeSWITCH).
   - Periodically executes balance debits on call at the beginning of debit interval.
   - Enforce call disconnection on insufficient balance.
   - Refunds the balance taken in advance at the call stop.

- POSTPAID
   - Executes balance debit on call-stop.

All call actions are logged into CGRateS's LogDB.

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   "stor_db" - (cdrDb)


2.1.4. DiameterAgent service
~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Responsible for the communication with Diameter server via diameter protocol.
Despite the name it is a flexible **Diameter Server**.

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   - none

- Config section in the CGRateS configuration file:
   - ``"diameter_agent": {...}``

2.1.5. CDR service
~~~~~~~~~~~~~~~~~~~
Centralized CDR server and CDR (raw or rated) **replicator**.

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   "stor_db" - (cdrDb)
   "data_db" - (accountDb)

- Config section in the CGRateS configuration file:
   - ``"cdrs": {...}``

2.1.6. CDRStats service
~~~~~~~~~~~~~~~~~~~~~~~
Computes real-time CDR stats. Capable with real-time fraud detection and mitigation with actions triggered.

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   "data_db"       - (dataDb)

- Config section in the CGRateS configuration file:
   - ``"cdrstats": {...}``

2.1.7. CDRC service
~~~~~~~~~~~~~~~~~~~
Gathers offline CDRs and post them to CDR Server - (CDRS component)

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   - none

- Config section in the CGRateS configuration file:
   - ``"cdrc": {...}``

2.1.8. Aliases service
~~~~~~~~~~~~~~~~~~~~~~~
Generic purpose **aliasing** system.

Possible applications:
   - Change destination name based on user or destination prefix matched.
   - Change lcr supplier name based on the user calling.
   - Locale specifics, ability to display specific tags in user defined language.

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   "data_db" - (accountDb)

- Config section in the CGRateS configuration file:
   - ``"aliases": {...}``

2.1.9. User service
~~~~~~~~~~~~~~~~~~~~
Generic purpose **user** system to maintain user profiles (LDAP similarity).

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   "data_db" - (accountDb)

- Config section in the CGRateS configuration file:
   - ``"users": {...}``

2.1.10. PubSub service
~~~~~~~~~~~~~~~~~~~~~~
PubSub service used to expose internal events to interested external components (eg: balance ops)

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   "data_db" - (accountDb)

- Config section in the CGRateS configuration file:
   - ``"pubsubs": {...}``


2.1.11. Resource Limiter service
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Resource Limiter service used to limit resources during authorization (eg: maximum calls per destination for an account)

- Communicates via:
   - RPC
   - internal/in-process *within the same running* **cgr-engine** process.

- Operates with the following CGRateS database(s): ::

   "data_db" - (accountDb)

- Config section in the CGRateS configuration file:
   - ``"rls": {...}``

2.1.12. APIER RPC service
~~~~~~~~~~~~~~~~~~~~~~~~~
RPC service used to expose external access towards internal components.

- Communicates via:
   - JSON/GOB over socket
   - JSON over HTTP
   - JSON over WebSocket

2.1.13. Cdre
~~~~~~~~~~~~
Component to retrieve rated CDRs from internal CDRs database.

- Communicates via:

- Operates with the following CGRateS database(s): ::

   "stor_db" - (cdrDb)

- Config section in the CGRateS configuration file:
   - ``"cdre": {...}``

2.1.14. Mailer
~~~~~~~~~~~~~~
TBD

- Communicates via:

- Operates with the following CGRateS database(s):

- Config section in the CGRateS configuration file:
   - ``"mailer": {...}``

2.1.15. Suretax
~~~~~~~~~~~~~~~
TBD

- Communicates via:

- Operates with the following CGRateS database(s):

- Config section in the CGRateS configuration file:
   - ``"suretax": {...}``


2.1.X Mediator service
~~~~~~~~~~~~~~~~~~~~~~

.. important:: This service is not valid anymore. Its functionality is replaced by CDRC and CDRS services.

Responsible to mediate the CDRs generated by Telecommunication Switch.

Has the ability to combine CDR fields into rating subject and run multiple mediation processes on the same record.

On Linux machines, able to work with inotify kernel subsystem in order to process the records close to real-time after the Switch has released them.


2.2. cgr-loader
---------------
Used for importing the rating information into the CGRateS database system.

Can be used to:
   - Import information from **csv files** to **data_db**.
   - Import information from **csv files** to **stor_db**. ``-to_stordb -tpid``
   - Import information from **stor_db** to **data_db**. ``-from_stordb -tpid``

::

 cgrates@OCS:~$ cgr-loader -help
 Usage of cgr-loader:
   -cdrstats_address string
         CDRStats service to contact for data reloads, empty to disable automatic data reloads (default "127.0.0.1:2013")
   -datadb_host string
         The DataDb host to connect to. (default "127.0.0.1")
   -datadb_name string
         The name/number of the DataDb to connect to. (default "11")
   -datadb_passwd string
         The DataDb user's password.
   -datadb_port string
         The DataDb port to bind to. (default "6379")
   -datadb_type string
         The type of the DataDb database <redis> (default "redis")
   -datadb_user string
         The DataDb user to sign in as.
   -dbdata_encoding string
         The encoding used to store object data in strings (default "msgpack")
   -disable_reverse_mappings
         Will disable reverse mappings rebuilding
   -dry_run
         When true will not save loaded data to dataDb but just parse it for consistency and errors.
   -flushdb
         Flush the database before importing
   -from_stordb
         Load the tariff plan from storDb to dataDb
   -migrate_rc8 string
         Migrate Accounts, Actions, ActionTriggers, DerivedChargers, ActionPlans and SharedGroups to RC8 structures, possible values: *all,acc,atr,act,dcs,apl,shg
   -path string
         The path to folder containing the data files (default "./")
   -rater_address string
         Rater service to contact for cache reloads, empty to disable automatic cache reloads (default "127.0.0.1:2013")
   -runid string
         Uniquely identify an import/load, postpended to some automatic fields
   -stats
         Generates statsistics about given data.
   -stordb_host string
         The storDb host to connect to. (default "127.0.0.1")
   -stordb_name string
         The name/number of the storDb to connect to. (default "cgrates")
   -stordb_passwd string
         The storDb user's password. (default "CGRateS.org")
   -stordb_port string
         The storDb port to bind to. (default "3306")
   -stordb_type string
         The type of the storDb database <mysql> (default "mysql")
   -stordb_user string
         The storDb user to sign in as. (default "cgrates")
   -timezone string
         Timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB> (default "Local")
   -to_stordb
         Import the tariff plan from files to storDb
   -validate
         When true will run various check on the loaded data to check for structural errors
   -verbose
         Enable detailed verbose logging output
   -version
         Prints the application version.


.. hint:: # cgr-loader -flushdb
.. hint:: # cgr-loader -verbose -datadb_port="27017" -datadb_type="mongo"

2.3. cgr-console
----------------
Command line tool used to interface with the RALs service. Able to execute **sub-commands**.

::

 cgrates@OCS:~$ cgr-console -help
 Usage of cgr-console:
   -rpc_encoding string
         RPC encoding used <gob|json> (default "json")
   -server string
         server address host:port (default "127.0.0.1:2012")
   -verbose
         Show extra info about command execution.
   -version
         Prints the application version.

 rif@grace:~$ cgr-console help_more
 2013/04/13 17:23:51
 Usage: cgr-console [cfg_opts...{-h}] <status|get_balance>

.. hint:: # cgr-console status

2.4. cgr-tester
---------------
Command line stress testing tool.

::

 cgrates@OCS:~$ cgr-tester --help
 Usage of cgr-tester:
  -datadb_host string
        The DataDb host to connect to. (default "127.0.0.1")
  -datadb_name string
        The name/number of the DataDb to connect to. (default "11")
  -datatdb_passwd string
        The DataDb user's password.
  -datadb_port string
        The DataDb port to bind to. (default "6379")
  -datadb_type string
        The type of the DataDb database <redis> (default "redis")
  -datadb_user string
        The DataDb user to sign in as.
  -category string
        The Record category to test. (default "call")
  -cpuprofile string
        write cpu profile to file
  -dbdata_encoding string
        The encoding used to store object data in strings. (default "msgpack")
  -destination string
        The destination to use in queries. (default "1002")
  -json
        Use JSON RPC
  -memprofile string
        write memory profile to this file
  -parallel int
        run n requests in parallel
  -rater_address string
        Rater address for remote tests. Empty for internal rater.
  -runs int
        stress cycle number (default 10000)
  -subject string
        The rating subject to use in queries. (default "1001")
  -tenant string
        The type of record to use in queries. (default "cgrates.org")
  -tor string
        The type of record to use in queries. (default "*voice")

.. hint:: # cgr-tester -runs=10000

2.5. cgr-migrator
-----------------
Command line migration tool.

::

 cgrates@OCS:~$ cgr-migrator --help
 Usage of cgr-migrator:
  -datadb_host string
      The DataDb host to connect to. (default "192.168.100.40")
  -datadb_name string
      The name/number of the DataDb to connect to. (default "10")
  -datadb_passwd string
      The DataDb user's password.
  -datadb_port string
      The DataDb port to bind to. (default "6379")
  -datadb_type string
      The type of the DataDb database <redis> (default "redis")
  -datadb_user string
      The DataDb user to sign in as. (default "cgrates")
  -dbdata_encoding string
      The encoding used to store object data in strings (default "msgpack")
  -dry_run
      When true will not save loaded data to dataDb but just parse it for consistency and errors.(default "false")
  -migrate string
      Fire up automatic migration *to use multiple values use ',' as separator 
      <*set_versions|*cost_details|*accounts|*actions|*action_triggers|*action_plans|*shared_groups> 
  -old_datadb_host string
      The DataDb host to connect to. (default "192.168.100.40")
  -old_datadb_name string
      The name/number of the DataDb to connect to. (default "10")
  -old_datadb_passwd string
      The DataDb user's password.
  -old_datadb_port string
      The DataDb port to bind to. (default "6379")
  -old_datadb_type string
      The type of the DataDb database <redis>
  -old_datadb_user string
      The DataDb user to sign in as. (default "cgrates")
  -old_dbdata_encoding string
      The encoding used to store object data in strings
  -old_stordb_host string
      The storDb host to connect to. (default "192.168.100.40")
  -old_stordb_name string
      The name/number of the storDb to connect to. (default "cgrates")
  -old_stordb_passwd string
      The storDb user's password.
  -old_stordb_port string
      The storDb port to bind to. (default "3306")
  -old_stordb_type string
      The type of the storDb database <mysql|postgres>
  -old_stordb_user string
      The storDb user to sign in as. (default "cgrates")
  -stats
      Generates statsistics about given data.(default "false")
  -stordb_host string
      The storDb host to connect to. (default "192.168.100.40")
  -stordb_name string
      The name/number of the storDb to connect to. (default "cgrates")
  -stordb_passwd string
      The storDb user's password.
  -stordb_port string
      The storDb port to bind to. (default "3306")
  -stordb_type string
      The type of the storDb database <mysql|postgres> (default "mysql")
  -stordb_user string
      The storDb user to sign in as. (default "cgrates")
  -verbose
      Enable detailed verbose logging output.(default "false")
  -version
      Prints the application version.
