OpenSIPS-CGRateS integration training
=====================================

Scenario:
---------

- Simple rating scenario: any destination dialed will be charged with a setup fee of 5 cents and 1 cent per second in 1 minute increment for the first minute then in second increments.
 - Cost simulation can be achieved via cost command: cgr-console 'cost Category="call" Tenant="cgrates.org" Subject="101" Destination="+4986517174963" TimeStart="2016-08-03T13:00:00Z" TimeEnd="2016-08-03T13:01:02Z"'

- Create 5 accounts (101, 102, 103, 104, 105).
 
 - Each account will receive 2 balances:
  - One *voice with 200 minutes on-net (destinations starting with 10 prefix)
  - One *monetary balance with 5 EUR/USD without destination limits

 - For each account we will set up 2 account triggers: one monitoring minimum balance (2 in our case) and second monitoring fraud (maximum balance) with threshold being set to 20. On thresholds being hit there will be a syslog warning

 - Account checks possible via account command: cgr-console 'accounts Tenant="cgrates.org" AccountIds=["101"]'

- Create one CDRStatS queue building up ASR, ACD, ACC, TCD, TCC metrics and having action triggers monitoring minimum asr of 35% - again with syslog as warning.
- CDRStatS queue metrics possible via console command: cgr-console 'cdrstats_metrics StatsQueueId="STATS_TEST"'



