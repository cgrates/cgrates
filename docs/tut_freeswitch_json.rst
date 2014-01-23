FreeSWITCH_ generating *http-json* CDRs
=======================================

Scenario
--------

- FreeSWITCH with *vanilla* configuration, replacing *mod_cdr_csv* with *mod_json_cdr*. 

 - Modified following users (with configs in *etc/freeswitch/directory/default*): 1001-prepaid, 1002-postpaid, 1003-pseudoprepaid, 1004-rated.
 - Have added inside default dialplan CGR own extensions just before routing towards users (*etc/freeswitch/dialplan/default.xml*).
 - FreeSWITCH configured to generate default *http-json* CDRs.

- **CGRateS** with following components:

 - CGR-SM started as prepaid controller, with debits taking place at 5s intervals.
 - CGR-Mediator compoenent attaching costs to the raw CDRs from FreeSWITCH_ inside CGR StorDB.
 - CGR-CDRE exporting mediated CDRs from CGR StorDB (export path: */tmp*).
 - CGR-History component keeping the archive of the rates modifications (path browsable with git client at */tmp/cgr_history*).


Starting FreeSWITCH_ with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/fs_json/freeswitch/etc/init.d/freeswitch start

To verify that FreeSWITCH_ is running we run the console command:

::

 fs_cli -x status


Starting **CGRateS** with custom configuration
----------------------------------------------

::

 /usr/share/cgrates/tutorials/fs_json/cgrates/etc/init.d/cgrates start

Check that cgrates is running

::

 cgr-console status


Loading **CGRateS** Tariff Plans
--------------------------------

For our tutorial we load again prepared data out of shared folder, containing following rules:

- Create the necessary timings (always, asap, peak, offpeak).
- Configure 3 destinations (1002, 1003 and 10 used as catch all rule).
- As rating we configure the following:

 - Rate id: *RT_10CNT* with connect fee of 20cents, 10cents per minute for the first 60s in 60s increments followed by 5cents per minute in 1s increments.
 - Rate id: *RT_20CNT* with connect fee of 40cents, 20cents per minute for the first 60s in 60s increments, followed by 10 cents per minute charged in 1s increments.
 - Rate id: *RT_40CNT* with connect fee of 80cents, 40cents per minute for the first 60s in 60s increments, follwed by 20cents per minute charged in 10s increments.
 - Will charge by default *RT_40CNT* for all FreeSWITCH_ destinations during peak times (Monday-Friday 08:00-19:00) and *RT_10CNT* during offpeatimes (rest).
 - Account 1001 will receive a special *deal* for 1002 and 1003 destinations during peak times with *RT_20CNT*, otherwise having default rating.

- Create 4 accounts (equivalent of 2 FreeSWITCH default test users - 1001, 1002, 1003, 1004).
- 1001, 1002, 1003, 1004 will receive 10units of *\*monetary* balance.
- For each balance created, attach 3 triggers to control the balance: log on balance=2, log on balance=20, log on 5 mins talked towards 10xx destination.

::

 cgr-loader -verbose -path=/usr/share/cgrates/tutorials/fs_json/cgrates/tariffplans

To verify that all actions successfully performed, we use following *cgr-console* commands:

- Make sure our rates were loaded successfully and they are already in cache:

 ::

  cgr-console get_cache_stats
  cgr-console get_cache_age 1002
  cgr-console get_cache_age RP_RETAIL1
  cgr-console get_cache_age *out:cgrates.org:call:*any
  cgr-console get_cache_age LOG_WARNING

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


1001 -> 1002
~~~~~~~~~~~~

Since the user 1001 is marked as *prepaid* inside FreeSWITCH_ directory configuration, calling between 1001 and 1002 should generate pre-auth and prepaid debits which can be checked with *get_balance* command integrated within *cgr-console* tool. Charging will be done based on time of day as described above.

*Note*: An important particularity to  note here is the ability of **CGRateS** SessionManager to refund units booked in advance (eg: if debit occurs every 10s and rate increments are set to 1s, the SessionManager will be smart enough to refund pre-booked credits for calls stoped in the middle of debit interval).

Check that 1001 balance is properly debitted, during the call:

::

 cgr-console get_balance cgrates.org 1001


1002 -> 1001
~~~~~~~~~~~~

The user 1002 is marked as *postpaid* inside FreeSWITCH_ hence his calls will be debited at the end of the call instead of during a call and his balance will be able to go on negative without influencing his new calls (no pre-auth).

To check that we had debits we use again console command, this time not during the call but at the end of it:

::

 cgr-console get_balance cgrates.org 1002


1003 -> 1001
~~~~~~~~~~~~

The user 1003 is marked as *pseudoprepaid* inside FreeSWITCH_ hence his calls will be considered same as prepaid (no call setups possible on negative balance due to pre-auth mechanism) but not handled automatically by session manager. His call costs will be calculated directly out of CDRs and balance updated by the time when mediation process occurs. This is sometimes a good compromise of prepaid running without influencing performance (there are no recurrent call debits during a call).

To check that there are no debits during or by the end of the call, but when the CDR reaches the CDRS component(which is close to real-time in case of *http-json* CDRs):

::

 cgr-console get_balance cgrates.org 1003


1004 -> 1001
~~~~~~~~~~~~

The user 1004 is marked as *rated* inside FreeSWITCH_ hence his calls not interact in any way with accounting subsystem. The only action perfomed by **CGRateS** related to his calls wil be rating/mediation of his CDRs.


Fraud detection
~~~~~~~~~~~~~~~

Since we have configured some action triggers (more than 20 units of balance topped-up or less than 2 and more than 5 units spent on *FS_USERS* we should be notified over syslog when things like unexpected events happen (eg: fraud with more than 20 units topped-up). To verify this mechanism simply add some random units into one account's balance:

::

 cgr-console add_balance cgrates.org 1003 21
 tail -f /var/log/syslog -n 20

*Note*: The actions are only executed once, in order to be repetive they need to be reset (via automated or manual process).


CDR processing
--------------

At the end of each call FreeSWITCH_ will issue a http post with the CDR. This will reach inside **CGRateS** through the *CDRS* component (close to real-time). Once in-there it will be instantly mediated and it is ready to be exported: 

::

 cgr-console export_cdrs csv


.. _FreeSWITCH: http://www.freeswitch.org/
.. _Jitsi: http://www.jitsi.org/
