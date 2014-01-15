Tutorial FS_JSON
================

Scenario:
---------

- Create the necessary timings (always, asap, peak, offpeak).
- Configure 3 destinations (1002, 1003 and 10 used as catch all rule).
- As rating we configure the following:

 - Rate id: *RT_10CNT* with connect fee of 20cents, 10cents per minute for the first 60s in 60s increments followed by 5cents per minute in 1s increments.
 - Rate id: *RT_20CNT* with connect fee of 40cents, 20cents per minute for the first 60s in 60s increments, followed by 10 cents per minute charged in 1s increments.
 - Rate id: *RT_40CNT* with connect fee of 80cents, 40cents per minute for the first 60s in 60s increments, follwed by 20cents per minute charged in 10s increments.
 - Will charge by default *RT_40CNT* for all FreeSWITCH_ destinations during peak times (Monday-Friday 08:00-19:00) and *RT_10CNT* during offpeatimes (rest).
 - Account 1001 will receive a special *deal* for 1002 and 1003 destinations during peak times with *RT_20CNT*, otherwise same as default rating.

- Create 4 accounts (equivalent of 2 FreeSWITCH default test users - 1001, 1002, 1003, 1004).
- 1001, 1002, 1003, 1004 will receive 10units of *\*monetary* balance.
- For each balance created, attach 3 triggers to control the balance: log on balance=2, log on balance=20, log on 5 mins talked towards 10xx destination.

