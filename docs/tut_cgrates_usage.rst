**CGRateS** Usage
=================

Loading **CGRateS** Tariff Plans
--------------------------------

Before proceeding to this step, you should have **CGRateS** installed and started with custom configuration, depending on the tutorial you have followed.

For our tutorial we load again prepared data out of shared folder, containing following rules:

- Create the necessary timings (always, asap, peak, offpeak).
- Configure 3 destinations (1002, 1003 and 10 used as catch all rule).
- As rating we configure the following:

 - Rate id: *RT_10CNT* with connect fee of 20cents, 10cents per minute for the first 60s in 60s increments followed by 5cents per minute in 1s increments.
 - Rate id: *RT_20CNT* with connect fee of 40cents, 20cents per minute for the first 60s in 60s increments, followed by 10 cents per minute charged in 1s increments.
 - Rate id: *RT_40CNT* with connect fee of 80cents, 40cents per minute for the first 60s in 60s increments, follwed by 20cents per minute charged in 10s increments.
 - Rate id: *RT_1CNT* having no connect fee and a rate of 1 cent per minute, chargeable in 1 minute increments.
 - Will charge by default *RT_40CNT* for all 10xx destinations during peak times (Monday-Friday 08:00-19:00) and *RT_10CNT* during offpeatimes (rest).
 - Account 1001 will receive a special *deal* for 1002 and 1003 destinations during peak times with *RT_20CNT*, otherwise having default rating.

- Accounting part will have following configured:
 - Create 5 accounts: 1001, 1002, 1003, 1004, 1007.
 - Create 1 account alias (1006 - alias of account 1002).
 - Create 1 rating profile alias (1006 - alias of rating profile 1001).
 - 1002, 1003, 1004 will receive 10units of *\*monetary* balance.
 - 1001 will receive 5 units of general  *\*monetary*, 5 units of shared balance in the shared group "SHARED_A" and 90 seconds of calling destination 1002 with special rates *RT_1CNT*.
 - 1007 will receive 0 units of shared balance in the shared group "SHARED_A".
 - Define the shared balance "SHARED_A" with debit policy *\*highest*.
 - For each balance created, attach 4 triggers to control the balance: log on balance<2, log on balance>20, log on 5 mins talked towards 10xx destination, disable the account and log if a balance is higher than 100 units.

- *DerivedCharging* will execute one extra mediation run when the sessions will have as account and rating subject 1001 resulting in a cloned session with most of parameters identical to original except RequestType which will be set on *rated* instead of original *prepaid* one. The extra run will be identified by *derived_run1* in CDRs.

- Will configure 4 extra CdrStatQueues:
 - *CDRST1* with 10 CDRs in the Queue and unlimited time window, calculating *ASR*, *ACD* and *ACC* for CDRs with Tenant matching *cgrates.org* and MediationRunId matching *default*. On this StatsQueue we will attach an ActionTrigger profile identified by *CDRST1_WARN*
 - *CDRST_1001* with 10 CDRs in the Queue and 10 minutes time window calculating *ASR*, *ACD* and *ACC* for CDRs with Tenant matching *cgrates.org*, RatingSubject matching *1001* and MediationRunId matching *default*. On this StatsQueue we will attach an ActionTrigger profile identified by *CDRST1001_WARN*
 - *CDRST_1002* with 10 CDRs in the Queue and 10 minutes time window calculating *ASR*, *ACD* and *ACC* for CDRs with Tenant matching *cgrates.org*, RatingSubject matching *1002* and MediationRunId matching *default*. On this StatsQueue we will attach an ActionTrigger profile identified by *CDRST1001_WARN*
 - *CDRST_1003* with 10 CDRs in the Queue and 10 minutes time window calculating *ASR*, and *ACD* for CDRs with Tenant matching *cgrates.org*, Destination matching *1003* and MediationRunId matching *default*. On this StatsQueue we will attach an ActionTrigger profile identified by *CDRST3_WARN*
 - The ActionTrigger *CDRST1_WARN* will monitor following StatsQueue Metric values:
  - ASR drop under 45 and a minimum of 3 CDRs in the StatsQueue will call Action profile *LOG_WARNING* which will log the StatsQueue to syslog. The Action will be recurrent with a sleep time of 1 minute.
  - ACD drop under 10 and a minimum of 5 CDRs in the StatsQueue will cause the same log to syslog. The Action will be recurrent with a sleep time of 1 minute.
  - ACC increase over 10 and a minimum of 5 CDRs in the StatsQueue will cause the StatsQueue to be again logged to syslog. The Action will be recurrent with a sleep time of 1 minute.

 - The ActionTrigger *CDRST1001_WARN* will monitor following StatsQueue Metric values:
  - ASR drop under 65 and a minimum of 3 CDRs in the StatsQueue will call Action profile *LOG_WARNING* which will log the StatsQueue to syslog. The Action will be recurrent with a sleep time of 1 minute.
  - ACD drop under 10 and a minimum of 5 CDRs in the StatsQueue will cause the same log to syslog. The Action will be recurrent with a sleep time of 1 minute.
  - ACC increase over 5 and a minimum of 5 CDRs in the StatsQueue will cause the StatsQueue to be again logged to syslog. The Action will be recurrent with a sleep time of 1 minute.

 - The ActionTrigger *CDRST3_WARN* will monitor ACD Metric and react at a minimum ACD of 60 with 5 CDRs in the StatsQueue by writing again to syslog. This ActionTrigger will be fired one time then cleared by the scheduler.

::

 cgr-loader -verbose -path=/usr/share/cgrates/tariffplans/tutorial

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
  cgr-console 'account Tenant="cgrates.org" Account="1005"'

- Query call costs so we can see our calls will have expected costs (final cost will result as sum of *ConnectFee* and *Cost* fields):

 ::

  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1002" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:00:20Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1002" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:01:25Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1003" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:00:20Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1003" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:01:25Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:00:20Z"'
  cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="1001" Destination="1004" TimeStart="2014-08-04T13:00:00Z" TimeEnd="2014-08-04T13:01:25Z"'

- Make sure *CDRStats Queues* were created:

 ::

  cgr-console cdrstats_queueids
  cgr-console 'cdrstats_metrics StatsQueueId="*default"'


Test calls
----------


1001 -> 1002
~~~~~~~~~~~~

Since the user 1001 is marked as *prepaid* inside the telecom switch, calling between 1001 and 1002 should generate pre-auth and prepaid debits which can be checked with *get_account* command integrated within *cgr-console* tool. Charging will be done based on time of day as described in the tariff plan definition above.

*Note*: An important particularity to  note here is the ability of **CGRateS** SessionManager to refund units booked in advance (eg: if debit occurs every 10s and rate increments are set to 1s, the SessionManager will be smart enough to refund pre-booked credits for calls stoped in the middle of debit interval).

Check that 1001 balance is properly deducted, during the call, and moreover considering that general balance has priority over the shared one debits for this call should take place at first out of general balance.

::

 cgr-console 'account Tenant="cgrates.org" Account="1001"'


1002 -> 1001
~~~~~~~~~~~~

The user 1002 is marked as *postpaid* inside the telecom switch hence his calls will be debited at the end of the call instead of during a call and his balance will be able to go on negative without influencing his new calls (no pre-auth).

To check that we had debits we use again console command, this time not during the call but at the end of it:

::

 cgr-console 'account Tenant="cgrates.org" Account="1002"'


1003 -> 1001
~~~~~~~~~~~~

The user 1003 is marked as *pseudoprepaid* inside the telecom switch hence his calls will be considered same as prepaid (no call setups possible on negative balance due to pre-auth mechanism) but not handled automatically by session manager. His call costs will be calculated directly out of CDRs and balance updated by the time when mediation process occurs. This is sometimes a good compromise of prepaid running without influencing performance (there are no recurrent call debits during a call).

To check that there are no debits during or by the end of the call, but when the CDR reaches the CDRS component(which is close to real-time in case of *http-json* CDRs):

::

 cgr-console 'account Tenant="cgrates.org" Account="1003"'


1004 -> 1001
~~~~~~~~~~~~

The user 1004 is marked as *rated* inside the telecom switch hence his calls not interact in any way with accounting subsystem. The only action perfomed by **CGRateS** related to his calls wil be rating/mediation of his CDRs.


1006 -> 1002
~~~~~~~~~~~~

Since the user 1006 is marked as *prepaid* inside the telecom switch, calling between 1006 and 1002 should generate pre-auth and prepaid debits which can be checked with *get_account* command integrated within *cgr-console* tool. One thing to note here is that 1006 is not defined as an account inside CGR Accounting Subsystem but as an alias of another account, hence *get_account* ran on 1006 will return "not found" and the debits can be monitored on the real account which is 1001.

Check that 1001 balance is properly debitted, during the call, and moreover considering that general balance has priority over the shared one debits for this call should take place at first out of general balance.

::

 cgr-console 'account Tenant="cgrates.org" Account="1006"'
 cgr-console 'account Tenant="cgrates.org" Account="1001"'


1007 -> 1002
~~~~~~~~~~~~

Since the user 1007 is marked as *prepaid* inside the telecom switch, calling between 1007 and 1002 should generate pre-auth and prepaid debits which can be checked with *get_account* command integrated within *cgr-console* tool. Since 1007 has no units left into his accounts but he has one balance marked as shared, debits for this call should take place in accounts which are a part of the same shared balance as the one of *1007/SHARED_A*, which in our scenario corresponds to the one of the account 1001.

Check that call can proceed even if 1007 has no units left into his own balances, and that the costs attached to the call towards 1002 are debited from the balance marked as shared within account 1001.

::

 cgr-console 'account Tenant="cgrates.org" Account="1007"'
 cgr-console 'account Tenant="cgrates.org" Account="1001"'


CDR Exporting
-------------

Once the CDRs are mediated, they are available to be exported. One can use available RPC APIs for that or directly call exports from console:

::

 cgr-console 'cdrs_export CdrFormat="csv" ExportDir="/tmp"'


Fraud detection
---------------

Since we have configured some action triggers (more than 20 units of balance topped-up or less than 2 and more than 5 units spent on *FS_USERS* we should be notified over syslog when things like unexpected events happen (eg: fraud with more than 20 units topped-up). Most important is the monitor for 100 units topped-up which will also trigger an account disable together with killing it's calls if prepaid debits are used.

To verify this mechanism simply add some random units into one account's balance:

::

 cgr-console 'balance_set Tenant="cgrates.org" Account="1003" Direction="*out" Value=23'
 tail -f /var/log/syslog -n 20

 cgr-console 'balance_set Tenant="cgrates.org" Account="1001" Direction="*out" Value=101'
 tail -f /var/log/syslog -n 20

On the CDRs side we will be able to integrate CdrStats monitors as part of our Fraud Detection system (eg: the increase of average cost for 1001 and 1002 accounts will signal us abnormalities, hence we will be notified via syslog).