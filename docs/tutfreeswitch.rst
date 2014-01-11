FreeSWITCH Integration Tutorials
================================

In these tutorials we exemplify few cases of integration between FreeSWITCH_ and CGRateS. We start with common steps, installation and postinstall processes then we dive in particular configurations, depending on the case we run.
 
Software installation
---------------------

As operating system we have choosen Debian Wheezy, since all the software components we use provide packaging for it.

Redis_
~~~~~~

Have installed Redis_ to serve as Rating and Accounting DB for CGRateS.

::

 apt-get install redis-server


MySQL_
~~~~~~

Have installed MySQL_ to serve as StorDB for CGRateS.

::

 apt-get install mysql-server

* To keep the tutorial simple, we have used as MySQL *root* password: *CGRateS.org*.


Git_
~~~~

Install Git_ used by CGRateS History Server as archiver.

::

 apt-get install git


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

At this point we have CGRateS installed but not yet configured. To facilitate the understanding and speed up the process, CGRateS comes already with the configurations used in this tutorial, available in the */usr/share/cgrates/tutorials* folder, so we will load them custom on each tutorial case.


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

Since by default FreeSWITCH_ restricts reading of the cdrs to it's own user/group, we will change permissions so we make sure *cgrates* user can read the folder also.

::
 chmod -R 754 /var/log/freeswitch/cdr-csv

Once installed we proceed with loading the configuration out of specific tutorial cases bellow.


Case 1: FreeSWITCH_ with *.csv* CDRs
------------------------------------

Scenario:
~~~~~~~~

* FreeSWITCH with default configuration. 

 * Modified following users: 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated, 1005-prepaid.
 * Have added inside default dialplan CGR own extensions just before routing towards users.
 * FreeSWITCH configured to generate default .csv CDRs, modified example template to add cgr_reqtype from user variables.

* CGRateS with following components:

 * CGR-SM started as prepaid controller.
 * CGR-CDRC component importing FreeSWITCH_ generated *.csv* CDRs into CGR.
 * CGR-CDRE exporting mediated CDRs


Starting FreeSWITCH_ with custom configuration:
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

 /usr/share/cgrates/tutorials/fs_csv/freeswitch/etc/init.d/freeswitch start

* To verify that FreeSWITCH_ is running we could run the console command:

::

 fs_cli -x status


Starting CGRateS with custom configuration:
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

 /usr/share/cgrates/tutorials/fs_csv/cgrates/etc/ini.d/cgrates start

* Check that cgrates is running

::
 cgr-console status


Loading CGRateS Tariff Plans
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

For our tutorial we load again prepared data out of shared folder, containing following rules:

* Create the necessary timings (always, asap).
* Configure 3 destinations (1002, 1003 and 10 used as catch all rule).
* As rating we configure the following:

 * Calls to 1002 destination will be rated with 20cents per minute for the first 60s in 60s increments then 10cents per minute in 1s increments.
 * Calls to 1003 destination will be rated with 40cents per minute for the first 60s in 30s increments then 20cents per minute in 10s increments.
 * Calls to other destinations (1001, 1004) will be rated with 10cents per minute for the first 60s(60s increments) then 5 cents per minute(1s increments).

* Create 4 accounts (equivalent of 2 FreeSWITCH default test users - 1001, 1002, 1003, 1004).
* 1001, 1002, 1003, 1004 will receive 10units of *\*monetary* balance.
* For each balance created, attach 3 triggers to control the balance: log on balance=2, log on balance=20, log on 5 mins talked towards 10xx destination.

::

 cgr-loader -verbose -path=/usr/share/cgrates/tutorials/fs_csv/cgrates/tariffplans


SIP UA - Jitsi_
---------------

On our ubuntu desktop host, we have installed Jitsi_ to be used as SIP UA, out of stable provided packages on `Jitsi download <https://jitsi.org/Main/Download>`_ and had Jitsi_ configured with 4 accounts out of default FreeSWITCH_ provided ones: 1001/CGRateS.org and 1002/CGRateS.org, 1003/CGRateS.org and 1004/CGRateS.org. For our tests we have configured 1001 as prepaid account, 1002 as postpaid, 1003 as pseudoprepaid and 1004 as rated, hence the type of charging will depend on the account calling.

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
 
