CGRateS Tutorial
================

Scenario:
---------

- Configure 3 destinations (1001,1002,1003).
- As rating we configure the following:

 - Rate id: *RT_10CNT* with connect fee of 20cents, 10cents per minute for the first 60s in 60s increments followed by 5cents per minute in 1s increments.
 - Rate id: *RT_20CNT* with connect fee of 40cents, 20cents per minute for the first 60s in 60s increments, followed by 10 cents per minute charged in 1s increments.
 - Rate id: *RT_40CNT* with connect fee of 80cents, 40cents per minute for the first 60s in 60s increments, follwed by 20cents per minute charged in 10s increments.
 - Rate id: *RT_1CNT* having no connect fee and a rate of 1 cent per minute, chargeable in 1 minute increments.
 - Rate id: *RT_1CNT_PER_SEC* having no connect fee and a rate of 1 cent per second, chargeable in 1 second increments.

- A call to destination 1003 will be automated closed after 12 seconds.

- Create 3 accounts (equivalent of FreeSWITCH default test users - 1001, 1002, 1003).
 
 - 1001, 1002,1003 will receive 10units of *monetary balance.


- Add 1 StatQueueProfile with 2 metrics :
 - *tcc total call cost 
 - *tcd total call duration 
 This will calculate these metrics based on FLTR_ACNT_1001_1002 (check if Account is 1001 or 1002 and RunID is *default)


- Add 2 ThresholdProfiles : 
 - THD_ACNT_1001 having as ActionIDs ACT_LOG_WARNING. THD_ACNT_1001 have MaxHits 1 so this will be executed once and after that will be deleted(the threshold not the profile).
 - THD_ACNT_1002 having as ActionIDs ACT_LOG_WARNING. THD_ACNT_1002 have MaxHits -1 so this will be executed each time when account 1002 make a call;