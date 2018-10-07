CGRateS Tutorial
================

Scenario:
---------

- Create the necessary timings (always, asap, peak, offpeak).
- Configure 3 destinations (1002, 1003 and 10 used as catch all rule).
- As rating we configure the following:

 - Rate id: *RT_10CNT* with connect fee of 20cents, 10cents per minute for the first 60s in 60s increments followed by 5cents per minute in 1s increments.
 - Rate id: *RT_20CNT* with connect fee of 40cents, 20cents per minute for the first 60s in 60s increments, followed by 10 cents per minute charged in 1s increments.
 - Rate id: *RT_40CNT* with connect fee of 80cents, 40cents per minute for the first 60s in 60s increments, follwed by 20cents per minute charged in 10s increments.
 - Will charge by default *RT_40CNT* for all FreeSWITCH_ destinations during peak time (Monday-Friday 08:00-19:00) and *RT_10CNT* during offpeatimes (rest).
 - Account 1001 will receive a special *deal* for 1002 and 1003 destinations during peak times with *RT_20CNT*, otherwise same as default rating.

- Create 5 accounts (equivalent of FreeSWITCH default test users - 1001, 1002, 1003, 1004, 1007).
 
 - 1002, 1003, 1004 will receive 10units of *monetary balance.
 - 1001 will receive 5 units of general *monetary and 5 units of shared balance in the shared group *SHARED_A*.
 - 1007 will receive 0 units of shared balance in the shared group *SHARED_A*.
 - Define the shared balance *SHARED_A* with debit policy *highest.

- Create 1 RatingProfile Alias: 1006 - alias of rating profile 1001.
- Create 1 Account Alias: 1006 - alias of account 1002.

- For each balance created, attach 3 triggers to control the balance: log on balance=2, log on balance=20, log on 5 mins talked towards 10xx destination.

- Add 4 CDRStats Queue configurations (extra to default one configured in .cfg file):

 - CDRStatsQueueId: *CDRST1* with purpose of monitoring calls on the Tenant=cgrates.org, and MediationRunId=default, with a QueueLength of 10 CDRs and no TimeWindow, monitoring ASR, ACD and ACC. Thrrough ActionTriggers *CDRST1_WARN* we monitor *min_asr(45), *min_acd(10) and *max_acc(10) parameters and log the warning on thresholds reached.
 - CDRStatsQueueId: *CDRST_1001* with a QueueLength of 10 CDRs and a TimeWindow of 10 minutes, gathering ASR,ACD and ACC metrics with CDR filters on Tenant=cgrates.org, Subject=1001 and MediationRunId=default with the purpose of monitoring calls of user 1001. To this queue we attach an ActionTrigger profile named *CDRST1001_WARN* which will monitor min_asr(65), min_acd(10), max_acc(5) and log to syslog a warning once thresholds are reached.

- ActionTrigger: *CDRST5_WARN* with thresholds explained bellow and having as action the *log on syslog

 - Threshold on *min_asr of 45, configured as recurrent with a sleep time of 1 minute and a minimum of 3 items in the stats queue.
 - Threshold on *min_acd of 10, configured as recurrent with a sleep time of 1 minute and a minimum of 5 items in the stats queue.
 - Threshold on *max_acc of 10, configured as recurrent with a sleep time of 1 minute and a minimum of 5 items in the stats queue.

- ActionTrigger: *CDRST6_WARN* with thresholds explained bellow and having as action the *log on syslog

 - Threshold on *min_asr of 30, configured as recurrent without sleep time and a minimum of 5 items in the stats queue.
 - Threshold on *min_acd of 3, configured as recurrent without sleep time and a minimum of 2 items in the stats queue.

