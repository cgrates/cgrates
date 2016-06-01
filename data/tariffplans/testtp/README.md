CGRateS - FSGermanyPrep1
==========================

Scenario:
---------

* Create the necessary timings (always, asap).
* Configure 1 destination: FS_USERS.
* Calls to FreeSWITCH users will be free and time independent.
* This rating profile will be valid for any rating subject.

* Create 5 prepaid accounts (equivalent of 5 FreeSWITCH default test users - 1001, 1002, 1003, 1004, 1005).
* Add to each of the accounts a monetary balance of 10 units.
* For each balance created, attach 3 triggers to control the balance: log on balance=2, log on balance=20, log on 15 mins talked towards FS_USERS destination.
