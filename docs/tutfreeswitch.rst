FreeSWITCH Integration Tutorials
================================

In these tutorials we exemplify few cases of integration between FreeSWITCH_ and CGRateS. We start with common steps, installation and postinstall processes then we dive in particular configurations, depending on the case we run.
 
Software installation
---------------------

As operating system we have choosen Debian Wheezy, since all the software components we use provide packaging for it.

Prerequisites:
~~~~~~~~~~~~~

Some components of CGRateS (whether enabled or not is up to the administrator) depend on external software like:

- Git_ used by CGRateS History Server as archiver.
- Redis_ to serve as Rating and Accounting DB for CGRateS.
- MySQL_ to serve as StorDB for CGRateS.

We will install them in one shoot using the command bellow.

::

 apt-get install git redis-server mysql-server

*Note*: For simplicity sake we have used as MySQL_ root password when asked: *CGRateS.org*.


FreeSWITCH_
~~~~~~~~~~~

More information regarding installing FreeSWITCH_ on Debian can be found on it's official `installation wiki <http://wiki.freeswitch.org/wiki/Installation_Guide#Debian_packages>`_.

To get FreeSWITCH_ installed and configured, we have choosen the simplest method, out of *vanilla* packages.

We got FreeSWITCH_ installed via following commands:

::

 gpg --keyserver pool.sks-keyservers.net --recv-key D76EDC7725E010CF
 gpg -a --export D76EDC7725E010CF | sudo apt-key add -
 cd /etc/apt/sources.list.d/
 wget http://apt.itsyscom.com/repos/apt/conf/freeswitch.apt.list
 apt-get update
 apt-get install freeswitch-meta-vanilla

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.


CGRateS
~~~~~~~

Installation steps are provided on CGRateS `install documentation <https://cgrates.readthedocs.org/en/latest/installation.html>`_.

To get CGRateS installed execute the following commands over ssh console:

::

 cd /etc/apt/sources.list.d/
 wget -O - http://apt.itsyscom.com/repos/apt/conf/cgrates.gpg.key|apt-key add -
 wget http://apt.itsyscom.com/repos/apt/conf/cgrates.apt.list
 apt-get update
 apt-get install cgrates

As described in post-install section, we will need to set up the MySQL_ database (using CGRateS.org as our root password):

::

 cd /usr/share/cgrates/storage/mysql/
 ./setup_cgr_db.sh root CGRateS.org localhost


Since by default FreeSWITCH_ restricts access to *.csv* CDRs to it's own user, we will add the *cgrates* user to freeswitch group.

::

 usermod -a -G freeswitch cgrates


At this point we have CGRateS installed but not yet configured. To facilitate the understanding and speed up the process, CGRateS comes already with the configurations used in this tutorial, available in the */usr/share/cgrates/tutorials* folder, so we will load them custom on each tutorial case.


SIP UA - Jitsi_
~~~~~~~~~~~~~~~

On our ubuntu desktop host, we have installed Jitsi_ to be used as SIP UA, out of stable provided packages on `Jitsi download <https://jitsi.org/Main/Download>`_ and had Jitsi_ configured with 4 accounts out of default FreeSWITCH_ provided ones: 1001/CGRateS.org and 1002/CGRateS.org, 1003/CGRateS.org and 1004/CGRateS.org.



Case 1: FreeSWITCH_ with *.csv* CDRs
------------------------------------

Scenario:
~~~~~~~~

- FreeSWITCH with *vanilla* configuration, minimal modifications to fit our needs. 

 - Modified following users (with configs in *etc/freeswitch/directory/default*): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated.
 - Have added inside default dialplan CGR own extensions just before routing towards users (*etc/freeswitch/dialplan/default.xml*).
 - FreeSWITCH configured to generate default .csv CDRs, modified example template to add cgr_reqtype from user variables (*etc/freeswitch/autoload_configs/cdr_csv.conf.xml*).

- CGRateS with following components:

 - CGR-SM started as prepaid controller, with debits taking place at 5s intervals.
 - CGR-CDRC component importing FreeSWITCH_ generated *.csv* CDRs into CGR (moving the processed *.csv* files to */tmp* folder).
 - CGR-Mediator compoenent attaching costs to the raw CDRs from CGR-CDRC.
 - CGR-CDRE exporting mediated CDRs (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path: */tmp/cgr_history*).


Starting FreeSWITCH_ with custom configuration:
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

 /usr/share/cgrates/tutorials/fs_csv/freeswitch/etc/init.d/freeswitch start

To verify that FreeSWITCH_ is running we run the console command:

::

 fs_cli -x status


Starting CGRateS with custom configuration:
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

 /usr/share/cgrates/tutorials/fs_csv/cgrates/etc/init.d/cgrates start

Check that cgrates is running

::

 cgr-console status


Loading CGRateS Tariff Plans
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

For our tutorial we load again prepared data out of shared folder, containing following rules:

- Create the necessary timings (always, asap).
- Configure 3 destinations (1002, 1003 and 10 used as catch all rule).
- As rating we configure the following:

 - Calls to 1002 destination will be rated with 20cents per minute for the first 60s in 60s increments then 10cents per minute in 1s increments.
 - Calls to 1003 destination will be rated with 40cents per minute for the first 60s in 30s increments then 20cents per minute in 10s increments.
 - Calls to other destinations (1001, 1004) will be rated with 10cents per minute for the first 60s(60s increments) then 5 cents per minute(1s increments).

- Create 4 accounts (equivalent of 2 FreeSWITCH default test users - 1001, 1002, 1003, 1004).
- 1001, 1002, 1003, 1004 will receive 10units of *\*monetary* balance.
- For each balance created, attach 3 triggers to control the balance: log on balance=2, log on balance=20, log on 5 mins talked towards 10xx destination.

::

 cgr-loader -verbose -path=/usr/share/cgrates/tutorials/fs_csv/cgrates/tariffplans

To verify that all actions successfully performed, we use following *cgr-console* commands:

- Make sure our rates were loaded successfully and they are already in cache:

 ::

  cgr-console get_cache_stats
  cgr-console get_cache_age 1002
  cgr-console get_cache_age RP_RETAIL
  cgr-console get_cache_age *out:cgrates.org:call:*any

- Make sure all our balances were topped-up:

 ::

  cgr-console get_balance cgrates.org 1001
  cgr-console get_balance cgrates.org 1002
  cgr-console get_balance cgrates.org 1003
  cgr-console get_balance cgrates.org 1004

- Query call costs so we can see our calls will have expected costs (final cost will result as sum of *ConnectFee* and *Cost* fields):

 ::

  cgr-console get_cost call cgrates.org 1001 1002 *now 20s
  cgr-console get_cost call cgrates.org 1001 1002 *now 1m25s
  cgr-console get_cost call cgrates.org 1001 1003 *now 20s
  cgr-console get_cost call cgrates.org 1001 1003 *now 1m25s
  cgr-console get_cost call cgrates.org 1001 1004 *now 20s
  cgr-console get_cost call cgrates.org 1001 1004 *now 1m25s


Test calls
----------

Calling between 1001 and 1003 should generate prepaid debits which can be checked with *get_balance* command integrated within *cgr-console* tool. The difference between calling from 1001 or 1003 should be reflected in fact that 1001 will generate real-time debits as opposite to 1003 which will only generate debits when CDRs will be processed. 

::
 cgr-console get_balance cgrates.org 1001
 cgr-console get_balance cgrates.org 1002


CDR processing
--------------

For every call FreeSWITCH_ will generate CDR records within the *Master.csv* file. In order to avoid double-processing them we will use the rotate mechanism built in FreeSWITCH_. We rotate files via *fs_console* command:

::

 fs_cli -x "cdr_csv rotate"

.. _Redis: http://redis.io/
.. _FreeSWITCH: http://www.freeswitch.org/
.. _MySQL: http://www.mysql.org/
.. _Jitsi: http://www.jitsi.org/
.. _Git: http://git-scm.com/

On each rotate CGR-CDRC component will be informed via *inotify* subsystem and will instantly process the CDR file. The records end up in CGRateS/StorDB inside *cdrs_primary* table via CGR-CDRS. Once in there mediation will occur, generating the costs inside *rated_cdrs* and *cost_details* tables.

Once the CDRs are mediated, can be exported as *.csv* format again via remote command offered by *cgr-console* tool:

::
 
