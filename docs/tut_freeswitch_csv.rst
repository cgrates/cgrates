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
 - CGR-CDRC component importing FreeSWITCH_ generated *.csv* CDRs into CGR and moving the processed *.csv* files to */tmp* folder.
 - CGR-Mediator compoenent attaching costs to the raw CDRs from CGR-CDRC inside CGR StorDB.
 - CGR-CDRE exporting mediated CDRs from CGR StorDB (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path browsable with git client at */tmp/cgr_history*).


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

  cgr-console cache_stats
  cgr-console 'cache_age 1002'
  cgr-console 'cache_age RP_RETAIL1'
  cgr-console 'cache_age *out:cgrates.org:call:*any'
  cgr-console 'cache_age LOG_WARNING'

- Make sure all our balances were topped-up:

 ::

  cgr-console 'account Tenant="cgrates.org" Account="1001"'
  cgr-console 'account Tenant="cgrates.org" Account="1002"'
  cgr-console 'account Tenant="cgrates.org" Account="1003"'
  cgr-console 'account Tenant="cgrates.org" Account="1004"'

- Query call costs so we can see our calls will have expected costs (final cost will result as sum of *ConnectFee* and *Cost* fields):

 ::

  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1002" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:00:20Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1002" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:01:25Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1003" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:00:20Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1003" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:01:25Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:00:20Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:01:25Z"'


Test calls
----------


1001 -> 1002
~~~~~~~~~~~~

Since the user 1001 is marked as *prepaid* inside FreeSWITCH_ directory configuration, calling between 1001 and 1002 should generate pre-auth and prepaid debits which can be checked with *get_balance* command integrated within *cgr-console* tool. As per our tariff plans, we should get first 60s charged as a whole, then in intervals of 1s (configured SessionManager debit interval of 10s).

*Note*: An important particularity to  note here is the ability of **CGRateS** SessionManager to refund units booked in advance (eg: if debit occurs every 10s and rate increments are set to 1s, the SessionManager will be smart enough to refund pre-booked credits for calls stoped in the middle of debit interval).

Check that 1001 balance is properly debitted, during the call:

::

 cgr-console 'account Tenant="cgrates.org" Account="1001"'


1002 -> 1001
~~~~~~~~~~~~

The user 1002 is marked as *postpaid* inside FreeSWITCH_ hence his calls will be debited at the end of the call instead of during a call and his balance will be able to go on negative without influencing his new calls (no pre-auth).

To check that we had debits we use again console command, this time not during the call but at the end of it:

::

 cgr-console 'account Tenant="cgrates.org" Account="1002"'


1003 -> 1001
~~~~~~~~~~~~

The user 1003 is marked as *pseudoprepaid* inside FreeSWITCH_ hence his calls will be considered same as prepaid (no call setups possible on negative balance due to pre-auth mechanism) but not handled automatically by session manager. His call costs will be calculated directly out of CDRs and balance updated by the time when mediation process occurs. This is sometimes a good compromise of prepaid running without influencing performance (there are no recurrent call debits during a call).

To check that there are no debits during or by the end of the call, but when the CDR is imported, run the command before and after rotating the FreeSWITCH_ *.csv* CDRs:

::

 cgr-console 'account Tenant="cgrates.org" Account="1003"'


1004 -> 1001
~~~~~~~~~~~~

The user 1004 is marked as *rated* inside FreeSWITCH_ hence his calls not interact in any way with accounting subsystem. The only action perfomed by **CGRateS** related to his calls wil be rating/mediation of his CDRs.


Fraud detection
~~~~~~~~~~~~~~~

Since we have configured some action triggers (more than 20 units of balance topped-up or less than 2 and more than 5 units spent on *FS_USERS* we should be notified over syslog when things like unexpected events happen (eg: fraud with more than 20 units topped-up). To verify this mechanism simply add some random units into one account's balance:

::

 cgr-console 'balance_set Tenant="cgrates.org" Account="1003" Direction="*out" Value=23'
 tail -f /var/log/syslog -n 20

*Note*: The actions are only executed once, in order to be repetive they need to be reset (via automated or manual process).


CDR processing
--------------

For every call FreeSWITCH_ will generate CDR records within the *Master.csv* file. 
In order to avoid double-processing we will use the rotate mechanism built in FreeSWITCH_. 
Once rotated, we will move the resulted files inside the path considered by **CGRateS** *CDRC* component as inbound.

These steps are automated in a script provided in the */usr/share/cgrates/scripts* location:

::

 /usr/share/cgrates/scripts/freeswitch_cdr_csv_rotate.sh


On each rotate CGR-CDRC component will be informed via *inotify* subsystem and will instantly process the CDR file. The records end up in **CGRateS**/StorDB inside *cdrs_primary* table via CGR-CDRS. As soon as the CDR will hit CDRS component, mediation will occur, either considering the costs calculated in case of prepaid and postpaid calls out of *cost_details* table or query it's own one from rater in case of *pseudoprepaid* and *rated* CDRs.

Once the CDRs are mediated, can be exported as *.csv* format again via remote command offered by *cgr-console* tool:

::

 cgr-console 'cdrs_export CdrFormat="csv" ExportDir="/tmp"'


.. _FreeSWITCH: http://www.freeswitch.org/
.. _Jitsi: http://www.jitsi.org/
