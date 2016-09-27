.. _cdrstats-main:

CDR Stats Server
================

Collects CDRs from various sources (eg: CGR-CDRS, CGR-Mediator, CGR-SM,
third-party CDR source via RPC) and builds real-time stats based on them. Each
StatsQueue has attached *ActionTriggers* with monitoring and actions
capabilities.


Principles of functionality:

- Standalone component (can be started individually on remote hardware, isolated form other CGRateS compoenents).
- Performance oriented. Should be able to process tens of thousands of CDRs per second.
- Cache driven technology. But SaveInterval can be set to store this information
  on redis.
- Stats are build within *StatsQueues* a CDR Stats Server being able to support
  unlimited number of StatsQueues. Each CDR will be passed to all of StatsQueues
  available and will be processed by individual StatsQueue based on configuration.
- Stats will be build inside Metrics (eg: ASR, ACD, ACC, TCC) and attached to specific StatsQueue.
- Each StatsQueue will have attached one *ActionTriggers* profile which will
  monitor Metrics values and react on thresholds reached (unlimited number of
  thresholds and reactions configurable).
- CDRs are processed by StatsQueues if they pass CDR field filters.
- CDRs are auto-removed from StatsQueues in a *fifo* manner if the QueueLength
  is reached or if they do not longer fit within TimeWindow defined.


Configuration
-------------

Individual StatsQueue configurations are loaded inside TariffPlan defitions, one
configuration object is internally represented as:

::

    type CdrStats struct {
        Id                string          // Config id, unique per config instance
        QueueLength       int             // Number of items in the stats buffer
        TimeWindow        time.Duration   // Will only keep the CDRs who's call setup time is not older than time.Now()-TimeWindow
        SaveInterval      time.Duration   // Interval to store the info into database
        Metrics           []string        // ASR, ACD, ACC, TCC, TCD, PDD
        SetupInterval     []time.Time     // CDRFieldFilter on SetupInterval, 2 or less items (>= start interval,< stop_interval)
        TOR               []string        // CDRFieldFilter on TORs
        CdrHost           []string        // CDRFieldFilter on CdrHosts
        CdrSource         []string        // CDRFieldFilter on CdrSources
        ReqType           []string        // CDRFieldFilter on ReqTypes
        Direction         []string        // CDRFieldFilter on Directions
        Tenant            []string        // CDRFieldFilter on Tenants
        Category          []string        // CDRFieldFilter on Categories
        Account           []string        // CDRFieldFilter on Accounts
        Subject           []string        // CDRFieldFilter on Subjects
        DestinationPrefix []string        // CDRFieldFilter on DestinationPrefixes
        UsageInterval     []time.Duration // CDRFieldFilter on UsageInterval, 2 or less items (>= Usage, <Usage)
        PddInterval       []time.Duration // CDRFieldFilter on PddInterval, 2 or less items (>= Pdd, <Pdd)
        Supplier          []string        // CDRFieldFilter on Suppliers
        DisconnectCause   []string        // Filter on DisconnectCause
        MediationRunIds   []string        // CDRFieldFilter on MediationRunIds
        RatedAccount      []string        // CDRFieldFilter on RatedAccounts
        RatedSubject      []string        // CDRFieldFilter on RatedSubjects
        CostInterval      []float64       // CDRFieldFilter on CostInterval, 2 or less items, (>=Cost, <Cost)
        Triggers          ActionTriggerPriotityList
    }


.. _cdrstats-metrics:

Metrics Types
-------------

- ACC *(Average Call Cost)*: Queue with the average call cost
- ACD *(Average Call Duration)*: Queue with the average call duration
- ASR *(Answer-Seizure Ratio)*: Queue with the answer ratio
- PDD *(Post Dial Delay )*: Queue with the average Post Dial Delay in seconds
- TCC *(Total Call Cost)*: Queue with the Total cost for the time frame.
- TCD *(Total Call Duration)*: Queue with the total call duration for the
time frame

ExternalQueries
---------------

The Metrics calculated are available to be real-time queried via RPC methods.

To facilitate interaction there are four commands built in the provided *cgr-console* tool:

- *cdrstats_queueids*: returns the queue ids processing CDR Stats.
- *cdrstats_metrics*: returns metrics calculated within specific CDRStatsQueue.
- *cdrstats_reload*: reloads the CdrStats configurations out of DataDb.
- *cdrstats_reset*: resets calculated metrics for one specific or all StatsQueues.


Example use
-----------

When you work with balance maybe you want to keep a eye in your users, so you
can add a new queue for the last 5 hours to check that your customer it's not
hacked, this is an example:

**CDR stats:**

.. code-block:: javascript

    "result":{
        "CdrStats": [
          {
            "Accounts": "my_account",
            "ActionTriggers": "FRAUD_CHECK",
            "Categories": "",
            "CdrHosts": "",
            "CdrSources": "",
            "CostInterval": "",
            "DestinationPrefixes": "",
            "Directions": "",
            "DisconnectCauses": "",
            "MediationRunIds": "",
            "Metrics": "TCC",
            "PddInterval": "",
            "QueueLength": "0",
            "RatedAccounts": "",
            "RatedSubjects": "",
            "ReqTypes": "",
            "SaveInterval": "15s",
            "SetupInterval": "",
            "Subjects": "",
            "Suppliers": "",
            "TORs": "",
            "Tenants": "foehn",
            "TimeWindow": "5h",
            "UsageInterval": ""
          }
        ],
        "CdrStatsId": "FRAUD_ACCOUNT",
        "TPid": "test"
      }



**Action Trigger:**

.. code-block:: javascript

    "result": {
        "ActionTriggers": [
          {
            "ActionsId": "LOG_WARNING",
            "BalanceCategory": "",
            "BalanceDestinationIds": "",
            "BalanceDirection": "",
            "BalanceExpirationDate": "",
            "BalanceId": "",
            "BalanceRatingSubject": "",
            "BalanceSharedGroup": "",
            "BalanceTimingTags": "",
            "BalanceType": "",
            "BalanceWeight": 0,
            "Id": "",
            "MinQueuedItems": 0,
            "MinSleep": "3h",
            "Recurrent": true,
            "ThresholdType": "\\*max_tcc",
            "ThresholdValue": 150,
            "Weight": 10
          }
        ],
        "ActionTriggersId": "FRAUD_CHECK",
        "TPid": "test"
      }

Using *cgr-console* you can check the status of the queue anytime:

.. code-block:: bash

    cgr-console 'cdrstats_metrics StatsQueueId="FRAUD_ACCOUNT"'
