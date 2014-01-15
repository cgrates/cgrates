Tutorial - FS_CSV
=================

Scenario:
---------

* Create the necessary timings (always, asap).
* Configure 1 destination: FS_USERS.
* Calls to FreeSWITCH users will be rated with 10cents per minute for the first 60s(60s increments) then 5 cents per minute(1s increments).
* This rating profile will be valid for any rating subject.

* Create 2 prepaid accounts (equivalent of 2 FreeSWITCH default test users - 1001, 1002).
* Add to each of the accounts a monetary balance of 10 units.
* For each balance created, attach 3 triggers to control the balance: log on balance=2, log on balance=20, log on 5 mins talked towards FS_USERS destination.
