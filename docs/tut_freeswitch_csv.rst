Case 1: FreeSWITCH_ generating *.csv* CDRs
==========================================

Scenario
--------

- FreeSWITCH with *vanilla* configuration, minimal modifications to fit our needs. 

 - Modified following users (with configs in *etc/freeswitch/directory/default*): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated.
 - Have added inside default dialplan CGR own extensions just before routing towards users (*etc/freeswitch/dialplan/default.xml*).
 - FreeSWITCH configured to generate default *.csv* CDRs, modified example template to add cgr_reqtype from user variables (*etc/freeswitch/autoload_configs/cdr_csv.conf.xml*).

- **CGRateS** with following components:

 - CGR-SM started as prepaid controller, with debits taking place at 5s intervals.
 - CGR-CDRC component importing FreeSWITCH_ generated *.csv* CDRs into CGR (moving the processed *.csv* files to */tmp* folder).
 - CGR-Mediator compoenent attaching costs to the raw CDRs from CGR-CDRC.
 - CGR-CDRE exporting mediated CDRs (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path: */tmp/cgr_history*).


Starting FreeSWITCH_ with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/fs_csv/freeswitch/etc/init.d/freeswitch start

To verify that FreeSWITCH_ is running we run the console command:

::

 fs_cli -x status


Starting **CGRateS** with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/fs_csv/cgrates/etc/init.d/cgrates start

Check that cgrates is running

::

 cgr-console status


Loading **CGRateS** Tariff Plans
--------------------------------

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


On each rotate CGR-CDRC component will be informed via *inotify* subsystem and will instantly process the CDR file. The records end up in **CGRateS**/StorDB inside *cdrs_primary* table via CGR-CDRS. Once in there mediation will occur, generating the costs inside *rated_cdrs* and *cost_details* tables.

Once the CDRs are mediated, can be exported as *.csv* format again via remote command offered by *cgr-console* tool:


